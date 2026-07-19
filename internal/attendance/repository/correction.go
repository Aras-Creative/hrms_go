package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/attendance/entity"
)

const (
	queryInsertCorrection = `
		INSERT INTO attendance_corrections (id, employee_id, date, clock_in, clock_out, status, reason, corrected_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	querySelectCorrection = `
		SELECT id, employee_id, date, clock_in, clock_out, status, reason, corrected_by, created_at
		FROM attendance_corrections
	`

	queryDeleteCorrection = `DELETE FROM attendance_corrections WHERE id = $1`

	queryUpdateCorrection = `
		UPDATE attendance_corrections
		SET clock_in = $2, clock_out = $3, status = $4, reason = $5
		WHERE id = $1
	`
)

var (
	queryCorrectionByID                  = querySelectCorrection + ` WHERE id = $1`
	queryCorrectionByEmployeeAndDate     = querySelectCorrection + ` WHERE employee_id = $1 AND date = $2`
	queryCorrectionPaginatedBase         = `SELECT ac.*, e.full_name AS employee_name FROM attendance_corrections ac JOIN employees e ON e.id = ac.employee_id`
)

type PostgresCorrectionRepo struct {
	db txContext
}

func NewPostgresCorrectionRepo(db *sqlx.DB) *PostgresCorrectionRepo {
	return &PostgresCorrectionRepo{db: db}
}

func (r *PostgresCorrectionRepo) WithTx(tx *sqlx.Tx) CorrectionRepository {
	return &PostgresCorrectionRepo{db: tx}
}

func (r *PostgresCorrectionRepo) Create(ctx context.Context, c *entity.AttendanceCorrection) error {
	_, err := r.db.ExecContext(ctx, queryInsertCorrection,
		c.ID, c.EmployeeID, c.Date, c.ClockIn, c.ClockOut, c.Status, c.Reason, c.CorrectedBy, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create correction: %w", err)
	}
	return nil
}

func (r *PostgresCorrectionRepo) Update(ctx context.Context, c *entity.AttendanceCorrection) error {
	result, err := r.db.ExecContext(ctx, queryUpdateCorrection,
		c.ID, c.ClockIn, c.ClockOut, c.Status, c.Reason,
	)
	if err != nil {
		return fmt.Errorf("failed to update correction: %w", err)
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

func (r *PostgresCorrectionRepo) FindByID(ctx context.Context, id string) (*entity.AttendanceCorrection, error) {
	var m CorrectionModel
	err := r.db.QueryRowxContext(ctx, queryCorrectionByID, id).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find correction: %w", err)
	}
	return modelToCorrection(&m), nil
}

func (r *PostgresCorrectionRepo) FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*entity.AttendanceCorrection, error) {
	var m CorrectionModel
	err := r.db.QueryRowxContext(ctx, queryCorrectionByEmployeeAndDate, employeeID, date).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find correction by employee and date: %w", err)
	}
	return modelToCorrection(&m), nil
}

func (r *PostgresCorrectionRepo) FindAllPaginated(ctx context.Context, searchName string, from, to time.Time, page, perPage int) ([]*CorrectionViewRow, int64, error) {
	where := " WHERE ac.date >= $1 AND ac.date <= $2"
	args := []interface{}{from, to}
	argIdx := 3

	if searchName != "" {
		where += fmt.Sprintf(" AND e.full_name ILIKE $%d", argIdx)
		args = append(args, "%"+searchName+"%")
		argIdx++
	}

	countQuery := `SELECT COUNT(*) FROM attendance_corrections ac JOIN employees e ON e.id = ac.employee_id` + where
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count corrections: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := queryCorrectionPaginatedBase + where + fmt.Sprintf(" ORDER BY ac.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list corrections: %w", err)
	}
	defer rows.Close()

	list := make([]*CorrectionViewRow, 0)
	for rows.Next() {
		var m CorrectionViewRow
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("failed to scan correction row: %w", err)
		}
		list = append(list, &m)
	}
	return list, total, rows.Err()
}

func (r *PostgresCorrectionRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, queryDeleteCorrection, id)
	if err != nil {
		return fmt.Errorf("failed to delete correction: %w", err)
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

func modelToCorrection(m *CorrectionModel) *entity.AttendanceCorrection {
	return entity.ReconstituteAttendanceCorrection(
		m.ID, m.EmployeeID, m.Date, m.ClockIn, m.ClockOut, m.Status, m.Reason, m.CorrectedBy, m.CreatedAt,
	)
}

var _ CorrectionRepository = (*PostgresCorrectionRepo)(nil)
