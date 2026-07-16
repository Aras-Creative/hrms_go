package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
)

const (
	queryInsertLeaveBalance = `
		INSERT INTO leave_balances (id, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	queryUpdateLeaveBalance = `
		UPDATE leave_balances SET total_days = $1, used_days = $2, updated_at = $3
		WHERE id = $4
	`
)

type PostgresLeaveBalanceRepo struct {
	db txContext
}

func NewPostgresLeaveBalanceRepo(db *sqlx.DB) *PostgresLeaveBalanceRepo {
	return &PostgresLeaveBalanceRepo{db: db}
}

func (r *PostgresLeaveBalanceRepo) WithTx(tx *sqlx.Tx) LeaveBalanceRepository {
	return &PostgresLeaveBalanceRepo{db: tx}
}

func (r *PostgresLeaveBalanceRepo) Create(ctx context.Context, lb *entity.LeaveBalance) error {
	_, err := r.db.ExecContext(ctx, queryInsertLeaveBalance,
		lb.ID, lb.EmployeeID, lb.LeaveTypeID, lb.Year, lb.TotalDays, lb.UsedDays, lb.CreatedAt, lb.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create leave balance: %w", err)
	}
	return nil
}

func (r *PostgresLeaveBalanceRepo) FindByEmployeeAndTypeYear(ctx context.Context, employeeID, leaveTypeID string, year int) (*entity.LeaveBalance, error) {
	var m LeaveBalanceModel
	err := r.db.QueryRowxContext(ctx, queryLeaveBalanceByEmpType, employeeID, leaveTypeID, year).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find leave balance: %w", err)
	}
	return modelToLeaveBalance(&m), nil
}

func (r *PostgresLeaveBalanceRepo) Update(ctx context.Context, lb *entity.LeaveBalance) error {
	result, err := r.db.ExecContext(ctx, queryUpdateLeaveBalance,
		lb.TotalDays, lb.UsedDays, lb.UpdatedAt, lb.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update leave balance: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("leave balance not found")
	}
	return nil
}

func (r *PostgresLeaveBalanceRepo) ConsumeBalance(ctx context.Context, id string, days int) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE leave_balances SET used_days = used_days + $1, updated_at = NOW() WHERE id = $2 AND used_days + $1 <= total_days`,
		days, id,
	)
	if err != nil {
		return fmt.Errorf("consume balance: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insufficient leave balance or balance not found")
	}
	return nil
}

var _ LeaveBalanceRepository = (*PostgresLeaveBalanceRepo)(nil)

func (r *PostgresLeaveBalanceRepo) FindAll(ctx context.Context, filter BalanceFilter) ([]*models.LeaveBalance, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	selectCols := `b.id, b.employee_id, b.leave_type_id, b.year, b.total_days, b.used_days,
		b.created_at, b.updated_at,
		e.full_name AS employee_name, e.employee_number, e.profile_photo_id,
		lt.name AS leave_type_name`

	fromClause := ` FROM leave_balances b
		LEFT JOIN employees e ON e.id = b.employee_id
		LEFT JOIN leave_types lt ON lt.id = b.leave_type_id`

	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.LeaveTypeID != "" {
		where = fmt.Sprintf(" WHERE b.leave_type_id = $%d", argIdx)
		args = append(args, filter.LeaveTypeID)
		argIdx++
	}
	if filter.Year > 0 {
		cond := fmt.Sprintf(" AND b.year = $%d", argIdx)
		if where == "" {
			where = " WHERE" + cond[4:]
		} else {
			where += cond
		}
		args = append(args, filter.Year)
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
		return nil, 0, fmt.Errorf("failed to count leave balances: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := `SELECT ` + selectCols + fromClause + where + fmt.Sprintf(" ORDER BY b.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	dataArgs := append(args, filter.PerPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list leave balances: %w", err)
	}
	defer rows.Close()

	list := make([]*models.LeaveBalance, 0)
	for rows.Next() {
		var m LeaveBalanceWithEmployeeModel
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("failed to scan leave balance: %w", err)
		}
		list = append(list, &models.LeaveBalance{
			ID:             m.ID,
			EmployeeID:     m.EmployeeID,
			EmployeeName:   m.EmployeeName,
			EmployeeNumber: m.EmployeeNumber,
			ProfilePhotoID: m.ProfilePhotoID,
			LeaveTypeID:    m.LeaveTypeID,
			LeaveTypeName:  m.LeaveTypeName,
			Year:           m.Year,
			TotalDays:      m.TotalDays,
			UsedDays:       m.UsedDays,
			CreatedAt:      m.CreatedAt,
			UpdatedAt:      m.UpdatedAt,
		})
	}
	return list, total, rows.Err()
}
