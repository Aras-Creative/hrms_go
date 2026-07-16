package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/schedule/entity"
)

const (
	queryInsertEmployeeWorkPattern = `
		INSERT INTO employee_work_patterns (id, employee_id, work_pattern_id, valid_from, valid_to, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	querySelectEmployeeWorkPattern = `
		SELECT id, employee_id, work_pattern_id, valid_from, valid_to, is_active, created_at, updated_at
		FROM employee_work_patterns
	`

	queryDeactivateCurrent = `
		UPDATE employee_work_patterns
		SET is_active = false, valid_to = $1, updated_at = NOW()
		WHERE employee_id = $2 AND is_active = true
	`

	queryUpsertEmployeeWorkPattern = `
		INSERT INTO employee_work_patterns (id, employee_id, work_pattern_id, valid_from, valid_to, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (employee_id, valid_from)
		DO UPDATE SET work_pattern_id = EXCLUDED.work_pattern_id,
		              is_active = EXCLUDED.is_active,
		              valid_to = EXCLUDED.valid_to,
		              updated_at = NOW()
	`
)

var (
	queryEWPByEmployeeAndDate = querySelectEmployeeWorkPattern + ` WHERE employee_id = $1 AND valid_from <= $2 AND (valid_to IS NULL OR valid_to >= $2) AND is_active = true ORDER BY valid_from DESC LIMIT 1`
	queryEWPActiveByEmployee  = querySelectEmployeeWorkPattern + ` WHERE employee_id = $1 AND is_active = true ORDER BY valid_from DESC LIMIT 1`
	queryEWPHistoryByEmployee = querySelectEmployeeWorkPattern + ` WHERE employee_id = $1 ORDER BY valid_from DESC`
)

type PostgresEmployeeWorkPatternRepo struct {
	db *sqlx.DB
}

func NewPostgresEmployeeWorkPatternRepo(db *sqlx.DB) *PostgresEmployeeWorkPatternRepo {
	return &PostgresEmployeeWorkPatternRepo{db: db}
}

func (r *PostgresEmployeeWorkPatternRepo) Create(ctx context.Context, ewp *entity.EmployeeWorkPattern) error {
	_, err := r.db.ExecContext(ctx, queryInsertEmployeeWorkPattern,
		ewp.ID, ewp.EmployeeID, ewp.WorkPatternID, ewp.ValidFrom, ewp.ValidTo, ewp.IsActive, ewp.CreatedAt, ewp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create employee work pattern: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeWorkPatternRepo) Upsert(ctx context.Context, ewp *entity.EmployeeWorkPattern) error {
	_, err := r.db.ExecContext(ctx, queryUpsertEmployeeWorkPattern,
		ewp.ID, ewp.EmployeeID, ewp.WorkPatternID, ewp.ValidFrom, ewp.ValidTo, ewp.IsActive, ewp.CreatedAt, ewp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert employee work pattern: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeWorkPatternRepo) FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*entity.EmployeeWorkPattern, error) {
	var m EmployeeWorkPatternModel
	err := r.db.QueryRowxContext(ctx, queryEWPByEmployeeAndDate, employeeID, date).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find employee work pattern: %w", err)
	}
	return modelToEWP(&m), nil
}

func (r *PostgresEmployeeWorkPatternRepo) FindActiveByEmployee(ctx context.Context, employeeID string) (*entity.EmployeeWorkPattern, error) {
	var m EmployeeWorkPatternModel
	err := r.db.QueryRowxContext(ctx, queryEWPActiveByEmployee, employeeID).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find active employee work pattern: %w", err)
	}
	return modelToEWP(&m), nil
}

func (r *PostgresEmployeeWorkPatternRepo) FindHistoryByEmployee(ctx context.Context, employeeID string) ([]*entity.EmployeeWorkPattern, error) {
	rows, err := r.db.QueryxContext(ctx, queryEWPHistoryByEmployee, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list employee work pattern history: %w", err)
	}
	defer rows.Close()

	var list []*entity.EmployeeWorkPattern
	for rows.Next() {
		var m EmployeeWorkPatternModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("failed to scan work pattern: %w", err)
		}
		list = append(list, modelToEWP(&m))
	}
	return list, rows.Err()
}

func (r *PostgresEmployeeWorkPatternRepo) DeactivateCurrent(ctx context.Context, employeeID string, validTo time.Time) error {
	_, err := r.db.ExecContext(ctx, queryDeactivateCurrent, validTo, employeeID)
	if err != nil {
		return fmt.Errorf("failed to deactivate current work pattern: %w", err)
	}
	return nil
}

func modelToEWP(m *EmployeeWorkPatternModel) *entity.EmployeeWorkPattern {
	return entity.ReconstituteEmployeeWorkPattern(
		m.ID, m.EmployeeID, m.WorkPatternID, m.ValidFrom, m.ValidTo, m.IsActive, m.CreatedAt, m.UpdatedAt,
	)
}
