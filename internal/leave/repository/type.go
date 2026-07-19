package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/leave/entity"
)

const queryInsertLeaveType = `
	INSERT INTO leave_types (id, name, default_days, is_paid, is_unlimited, is_half_day, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type PostgresLeaveTypeRepo struct {
	db *sqlx.DB
}

func NewPostgresLeaveTypeRepo(db *sqlx.DB) *PostgresLeaveTypeRepo {
	return &PostgresLeaveTypeRepo{db: db}
}

func (r *PostgresLeaveTypeRepo) Create(ctx context.Context, lt *entity.LeaveType) error {
	_, err := r.db.ExecContext(ctx, queryInsertLeaveType,
		lt.ID,          // $1
		lt.Name,        // $2
		lt.DefaultDays, // $3
		lt.IsPaid,      // $4
		lt.IsUnlimited, // $5
		lt.IsHalfDay,   // $6
		lt.IsActive,    // $7
		lt.CreatedAt,   // $8
		lt.UpdatedAt,   // $9
	)
	if err != nil {
		return fmt.Errorf("failed to create leave type: %w", err)
	}
	return nil
}

func (r *PostgresLeaveTypeRepo) FindByID(ctx context.Context, id string) (*entity.LeaveType, error) {
	var m LeaveTypeModel
	err := r.db.QueryRowxContext(ctx, queryLeaveTypeByID, id).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find leave type: %w", err)
	}
	return modelToLeaveType(&m), nil
}

func (r *PostgresLeaveTypeRepo) FindAllActive(ctx context.Context) ([]*entity.LeaveType, error) {
	rows, err := r.db.QueryxContext(ctx, queryLeaveTypesAllActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list active leave types: %w", err)
	}
	defer rows.Close()

	var list []*entity.LeaveType
	for rows.Next() {
		var m LeaveTypeModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("failed to scan leave type: %w", err)
		}
		list = append(list, modelToLeaveType(&m))
	}
	return list, rows.Err()
}

var _ LeaveTypeRepository = (*PostgresLeaveTypeRepo)(nil)

func (r *PostgresLeaveTypeRepo) Update(ctx context.Context, lt *entity.LeaveType) error {
	result, err := r.db.ExecContext(ctx, queryUpdateLeaveType,
		lt.Name, lt.DefaultDays, lt.IsPaid, lt.IsUnlimited, lt.IsHalfDay, lt.IsActive, lt.UpdatedAt, lt.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update leave type: %w", err)
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
