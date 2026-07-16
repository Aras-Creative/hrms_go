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
	queryUpsertOverride = `
		INSERT INTO employee_schedule_overrides (id, employee_id, date, is_working_day, start_time, end_time, reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (employee_id, date) DO UPDATE SET
			is_working_day = EXCLUDED.is_working_day,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			reason = EXCLUDED.reason,
			updated_at = EXCLUDED.updated_at
	`

	querySelectOverride = `
		SELECT id, employee_id, date, is_working_day, start_time, end_time, reason, created_at, updated_at
		FROM employee_schedule_overrides
	`

	queryDeleteOverride      = `DELETE FROM employee_schedule_overrides WHERE id = $1`
	queryDeleteOverrideByEmp = `DELETE FROM employee_schedule_overrides WHERE employee_id = $1 AND date = $2`
)

var (
	queryOverrideByID        = querySelectOverride + ` WHERE id = $1`
	queryOverridesByEmpRange = querySelectOverride + ` WHERE employee_id = $1 AND date >= $2 AND date <= $3 ORDER BY date ASC`
	queryOverridesByRange    = querySelectOverride + ` WHERE date >= $1 AND date <= $2 ORDER BY employee_id, date ASC`
)

type PostgresScheduleOverrideRepo struct {
	db *sqlx.DB
}

func NewPostgresScheduleOverrideRepo(db *sqlx.DB) *PostgresScheduleOverrideRepo {
	return &PostgresScheduleOverrideRepo{db: db}
}

func (r *PostgresScheduleOverrideRepo) Upsert(ctx context.Context, o *entity.EmployeeScheduleOverride) error {
	_, err := r.db.ExecContext(ctx, queryUpsertOverride,
		o.ID, o.EmployeeID, o.Date, o.IsWorkingDay, o.StartTime, o.EndTime, o.Reason, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert schedule override: %w", err)
	}
	return nil
}

func (r *PostgresScheduleOverrideRepo) FindByID(ctx context.Context, id string) (*entity.EmployeeScheduleOverride, error) {
	var m EmployeeScheduleOverrideModel
	err := r.db.QueryRowxContext(ctx, queryOverrideByID, id).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find schedule override: %w", err)
	}
	return modelToOverride(&m), nil
}

func (r *PostgresScheduleOverrideRepo) FindByEmployeeAndDateRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.EmployeeScheduleOverride, error) {
	rows, err := r.db.QueryxContext(ctx, queryOverridesByEmpRange, employeeID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule overrides: %w", err)
	}
	defer rows.Close()
	return scanOverrides(rows)
}

func (r *PostgresScheduleOverrideRepo) FindByDateRange(ctx context.Context, from, to time.Time) ([]*entity.EmployeeScheduleOverride, error) {
	rows, err := r.db.QueryxContext(ctx, queryOverridesByRange, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule overrides by date: %w", err)
	}
	defer rows.Close()
	return scanOverrides(rows)
}

var ErrOverrideNotFound = fmt.Errorf("schedule override not found")

func (r *PostgresScheduleOverrideRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, queryDeleteOverride, id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule override: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrOverrideNotFound
	}
	return nil
}

func (r *PostgresScheduleOverrideRepo) DeleteByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) error {
	_, err := r.db.ExecContext(ctx, queryDeleteOverrideByEmp, employeeID, date)
	if err != nil {
		return fmt.Errorf("failed to delete schedule override: %w", err)
	}
	return nil
}

func (r *PostgresScheduleOverrideRepo) DeleteFutureOverridesByEmployee(ctx context.Context, employeeID string, since time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE daily_attendances SET schedule_override_id = NULL WHERE employee_id = $1 AND date >= $2 AND schedule_override_id IS NOT NULL`, employeeID, since)
	if err != nil {
		return fmt.Errorf("failed to nullify schedule_override_id in daily_attendances: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM employee_schedule_overrides WHERE employee_id = $1 AND date >= $2`, employeeID, since)
	if err != nil {
		return fmt.Errorf("failed to delete future schedule overrides: %w", err)
	}
	return nil
}

func scanOverrides(rows *sqlx.Rows) ([]*entity.EmployeeScheduleOverride, error) {
	var list []*entity.EmployeeScheduleOverride
	for rows.Next() {
		var m EmployeeScheduleOverrideModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}
		list = append(list, modelToOverride(&m))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func modelToOverride(m *EmployeeScheduleOverrideModel) *entity.EmployeeScheduleOverride {
	return entity.ReconstituteEmployeeScheduleOverride(
		m.ID, m.EmployeeID, m.Date, m.IsWorkingDay, m.StartTime, m.EndTime, m.Reason, m.CreatedAt, m.UpdatedAt,
	)
}

const queryScheduleOverview = `
	SELECT
		e.id AS employee_id,
		e.full_name,
		e.employee_number,
		d.name AS designation_name,
		ewp.work_pattern_id,
		wp.name AS working_pattern_name,
		date_series.d::date AS date,
		EXTRACT(DOW FROM date_series.d)::int AS day_of_week,
		wpd.id AS pattern_detail_id,
		eso.id AS override_id,
		eso.is_working_day AS override_is_working,
		eso.start_time AS override_start_time,
		eso.end_time AS override_end_time,
		eso.reason AS override_notes,
		wpd.start_time AS pattern_start_time,
		wpd.end_time AS pattern_end_time,
		wpd.working_type AS pattern_working_type
	FROM generate_series($1::date, $2::date, '1 day'::interval) AS date_series(d)
	CROSS JOIN employees e
	LEFT JOIN designations d ON d.id = e.designation_id
	LEFT JOIN employee_work_patterns ewp ON ewp.employee_id = e.id
		AND ewp.valid_from <= date_series.d::date
		AND (ewp.valid_to IS NULL OR ewp.valid_to >= date_series.d::date)
		AND ewp.is_active = true
	LEFT JOIN work_patterns wp ON wp.id = ewp.work_pattern_id AND wp.is_active = true
	LEFT JOIN work_pattern_details wpd ON wpd.work_pattern_id = wp.id
		AND wpd.day_of_week = EXTRACT(DOW FROM date_series.d)::int
	LEFT JOIN employee_schedule_overrides eso ON eso.employee_id = e.id
		AND eso.date = date_series.d::date
	WHERE ($3 = '' OR e.id::text = $3)
		AND ($4 = '' OR e.designation_id::text = $4)
		AND ($5 = '' OR e.full_name ILIKE '%' || $5 || '%' OR e.employee_number ILIKE '%' || $5 || '%')
	ORDER BY e.id, date_series.d
`

func (r *PostgresScheduleOverrideRepo) QueryOverview(ctx context.Context, p ScheduleOverviewParams) ([]ScheduleOverviewRow, error) {
	rows, err := r.db.QueryxContext(ctx, queryScheduleOverview,
		p.From, p.To, p.EmployeeID, p.DesignationID, p.Search,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedule overview: %w", err)
	}
	defer rows.Close()

	var list []ScheduleOverviewRow
	for rows.Next() {
		var row ScheduleOverviewRow
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("failed to scan overview row: %w", err)
		}
		list = append(list, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}
