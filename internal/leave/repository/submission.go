package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
)

const (
	queryInsertSubmission = `
		INSERT INTO leave_submissions (
			id, employee_id, leave_type_id, start_date, end_date, days,
			reason, attachment_id, status, approved_by, approved_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	querySelectSubmission = `
		SELECT id, employee_id, leave_type_id, start_date, end_date, days,
			reason, attachment_id, status, approved_by, approved_at,
			created_at, updated_at
		FROM leave_submissions
	`

	queryUpdateSubmission = `
		UPDATE leave_submissions SET
			status = $1, approved_by = $2, approved_at = $3, updated_at = $4
		WHERE id = $5
	`

	queryCheckOverlap = `
		SELECT EXISTS (
			SELECT 1 FROM leave_submissions
			WHERE employee_id = $1
			  AND status IN ('pending', 'approved')
			  AND start_date <= $3
			  AND end_date >= $2
			  AND ($4 = '' OR id != $4::uuid)
		)
	`
)

var (
	querySubmissionByID        = querySelectSubmission + ` WHERE id = $1`
	querySubmissionsByEmployee = querySelectSubmission + ` WHERE employee_id = $1`
)

type PostgresLeaveSubmissionRepo struct {
	db txContext
}

func NewPostgresLeaveSubmissionRepo(db *sqlx.DB) *PostgresLeaveSubmissionRepo {
	return &PostgresLeaveSubmissionRepo{db: db}
}

func (r *PostgresLeaveSubmissionRepo) WithTx(tx *sqlx.Tx) LeaveSubmissionRepository {
	return &PostgresLeaveSubmissionRepo{db: tx}
}

var _ LeaveSubmissionRepository = (*PostgresLeaveSubmissionRepo)(nil)

func (r *PostgresLeaveSubmissionRepo) Create(ctx context.Context, s *entity.LeaveSubmission) error {
	_, err := r.db.ExecContext(ctx, queryInsertSubmission,
		s.ID, s.EmployeeID, s.LeaveTypeID, s.StartDate, s.EndDate, s.Days,
		s.Reason, s.AttachmentID, string(s.Status), s.ApprovedBy, s.ApprovedAt,
		s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create leave submission: %w", err)
	}
	return nil
}

func (r *PostgresLeaveSubmissionRepo) FindByID(ctx context.Context, id string) (*entity.LeaveSubmission, error) {
	var m LeaveSubmissionModel
	err := r.db.QueryRowxContext(ctx, querySubmissionByID, id).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find leave submission: %w", err)
	}
	return modelToSubmission(&m), nil
}

func (r *PostgresLeaveSubmissionRepo) FindByEmployeeID(ctx context.Context, filter LeaveSubmissionFilter) ([]*entity.LeaveSubmission, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	baseArgs := []interface{}{filter.EmployeeID}
	whereExtra := ""
	nextIdx := 2

	if filter.Status != "" {
		whereExtra = fmt.Sprintf(" AND status = $%d", nextIdx)
		baseArgs = append(baseArgs, filter.Status)
		nextIdx++
	}

	countQuery := `SELECT COUNT(*) FROM leave_submissions WHERE employee_id = $1` + whereExtra
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, baseArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to count submissions: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage

	dataQuery := querySubmissionsByEmployee + whereExtra + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", nextIdx, nextIdx+1)
	dataArgs := append(baseArgs, filter.PerPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list submissions: %w", err)
	}
	defer rows.Close()

	list := make([]*entity.LeaveSubmission, 0)
	for rows.Next() {
		var m LeaveSubmissionModel
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("failed to scan submission: %w", err)
		}
		list = append(list, modelToSubmission(&m))
	}
	return list, total, rows.Err()
}

func (r *PostgresLeaveSubmissionRepo) HasOverlap(ctx context.Context, employeeID string, startDate, endDate time.Time, excludeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, queryCheckOverlap, employeeID, startDate, endDate, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check overlap: %w", err)
	}
	return exists, nil
}

func (r *PostgresLeaveSubmissionRepo) HasApprovedLeave(ctx context.Context, employeeID string, date time.Time) (bool, *string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `
		SELECT lt.name FROM leave_submissions ls
		JOIN leave_types lt ON lt.id = ls.leave_type_id
		WHERE ls.employee_id = $1 AND ls.status = 'approved'
			AND ls.start_date <= $2::date AND ls.end_date >= $2::date
			AND lt.is_half_day = false
		LIMIT 1
	`, employeeID, date).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to check approved leave: %w", err)
	}
	return true, &name, nil
}

func (r *PostgresLeaveSubmissionRepo) FindAll(ctx context.Context, filter SubmissionFilter) ([]*models.LeaveSubmission, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	selectCols := `s.id, s.employee_id, s.leave_type_id, s.start_date, s.end_date, s.days,
		s.reason, s.attachment_id, s.status, s.approved_by, s.approved_at,
		s.created_at, s.updated_at,
		e.full_name AS employee_name, e.employee_number, e.profile_photo_id,
		lt.name AS leave_type_name`

	fromClause := ` FROM leave_submissions s 
		LEFT JOIN employees e ON e.id = s.employee_id
		LEFT JOIN leave_types lt ON lt.id = s.leave_type_id`

	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.Status != "" {
		where = fmt.Sprintf(" WHERE s.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.StartDate != nil {
		cond := fmt.Sprintf(" AND s.start_date >= $%d", argIdx)
		if where == "" {
			where = " WHERE" + cond[4:]
		} else {
			where += cond
		}
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		cond := fmt.Sprintf(" AND s.end_date <= $%d", argIdx)
		if where == "" {
			where = " WHERE" + cond[4:]
		} else {
			where += cond
		}
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if filter.Search != "" {
		cond := fmt.Sprintf(" AND e.full_name ILIKE $%d", argIdx)
		if where == "" {
			where = " WHERE" + cond[4:]
		} else {
			where += cond
		}
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := `SELECT COUNT(*)` + fromClause + where
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count all submissions: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := `SELECT ` + selectCols + fromClause + where + fmt.Sprintf(" ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	dataArgs := append(args, filter.PerPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all submissions: %w", err)
	}
	defer rows.Close()

	list := make([]*models.LeaveSubmission, 0)
	for rows.Next() {
		var m LeaveSubmissionWithEmployeeModel
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("failed to scan submission: %w", err)
		}
		list = append(list, &models.LeaveSubmission{
			ID:             m.ID,
			EmployeeID:     m.EmployeeID,
			LeaveTypeID:    m.LeaveTypeID,
			LeaveTypeName:  m.LeaveTypeName,
			StartDate:      m.StartDate,
			EndDate:        m.EndDate,
			Days:           m.Days,
			Reason:         m.Reason,
			AttachmentID:   m.AttachmentID,
			Status:         m.Status,
			ApprovedBy:     m.ApprovedBy,
			ApprovedAt:     m.ApprovedAt,
			CreatedAt:      m.CreatedAt,
			UpdatedAt:      m.UpdatedAt,
			EmployeeName:   m.EmployeeName,
			EmployeeNumber: m.EmployeeNumber,
			ProfilePhotoID: m.ProfilePhotoID,
		})
	}
	return list, total, rows.Err()
}

func (r *PostgresLeaveSubmissionRepo) Update(ctx context.Context, s *entity.LeaveSubmission) error {
	result, err := r.db.ExecContext(ctx, queryUpdateSubmission,
		string(s.Status), s.ApprovedBy, s.ApprovedAt, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update submission: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return nil
	}
	return nil
}

func modelToSubmission(m *LeaveSubmissionModel) *entity.LeaveSubmission {
	status, _ := entity.ParseLeaveStatus(m.Status)
	return entity.ReconstituteLeaveSubmission(
		m.ID, m.EmployeeID, m.LeaveTypeID,
		m.StartDate, m.EndDate, m.Days, m.Reason, m.AttachmentID,
		status, m.ApprovedBy, m.ApprovedAt, m.CreatedAt, m.UpdatedAt,
	)
}
