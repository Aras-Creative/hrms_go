package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/attendance/entity"
)

const (
	queryInsertPunch = `
		INSERT INTO punches (id, employee_id, type, timestamp, date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	querySelectPunch = `
		SELECT id, employee_id, type, timestamp, date, created_at
		FROM punches
	`
)

var (
	queryPunchByEmployeeRange = querySelectPunch + ` WHERE employee_id = $1 AND timestamp >= $2 AND timestamp <= $3 ORDER BY timestamp ASC`
	queryPunchTodayByEmployee = querySelectPunch + ` WHERE employee_id = $1 AND date = CURRENT_DATE ORDER BY timestamp ASC`
)

type PostgresPunchRepo struct {
	db *sqlx.DB
}

func NewPostgresPunchRepo(db *sqlx.DB) *PostgresPunchRepo {
	return &PostgresPunchRepo{db: db}
}

func (r *PostgresPunchRepo) Create(ctx context.Context, p *entity.Punch) error {
	_, err := r.db.ExecContext(ctx, queryInsertPunch,
		p.ID, p.EmployeeID, string(p.Type), p.Timestamp, p.Date, p.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create punch: %w", err)
	}
	return nil
}

func (r *PostgresPunchRepo) FindByEmployeeAndDateRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.Punch, error) {
	rows, err := r.db.QueryxContext(ctx, queryPunchByEmployeeRange, employeeID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list punches: %w", err)
	}
	defer rows.Close()
	return scanPunches(rows)
}

func (r *PostgresPunchRepo) FindTodayByEmployee(ctx context.Context, employeeID string) ([]*entity.Punch, error) {
	rows, err := r.db.QueryxContext(ctx, queryPunchTodayByEmployee, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list today punches: %w", err)
	}
	defer rows.Close()
	return scanPunches(rows)
}

func scanPunches(rows *sqlx.Rows) ([]*entity.Punch, error) {
	var list []*entity.Punch
	for rows.Next() {
		var m PunchModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("failed to scan punch: %w", err)
		}
		list = append(list, modelToPunch(&m))
	}
	return list, rows.Err()
}

func modelToPunch(m *PunchModel) *entity.Punch {
	return entity.ReconstitutePunch(
		m.ID, m.EmployeeID, entity.PunchType(m.Type), m.Timestamp, m.Date, m.CreatedAt,
	)
}

var _ PunchRepository = (*PostgresPunchRepo)(nil)
