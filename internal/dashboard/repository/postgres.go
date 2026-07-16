package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PostgresDashboardRepo struct {
	db *sqlx.DB
}

func NewPostgresDashboardRepo(db *sqlx.DB) *PostgresDashboardRepo {
	return &PostgresDashboardRepo{db: db}
}

const queryResolvedStatus = `
	WITH resolved AS (
		SELECT
			d::date AS day,
			e.id AS employee_id,
			COALESCE(da.status,
				CASE
					WHEN eso.id IS NOT NULL AND eso.is_working_day = false THEN 'day_off'
					WHEN eso.id IS NOT NULL AND eso.start_time IS NOT NULL AND eso.start_time != '' THEN
						CASE WHEN eso.end_time IS NOT NULL AND eso.end_time != '' AND NOW() > (d::date + eso.end_time::time) THEN 'absent' ELSE 'no_punch' END
					WHEN wpd.id IS NOT NULL AND wpd.working_type = 'dynamic' THEN 'no_punch'
					WHEN wpd.id IS NOT NULL AND wpd.start_time IS NOT NULL AND wpd.start_time != ''
						AND (wpd.working_type IS NULL OR wpd.working_type != 'off') THEN
						CASE WHEN wpd.end_time IS NOT NULL AND wpd.end_time != '' AND NOW() > (d::date + wpd.end_time::time) THEN 'absent' ELSE 'no_punch' END
					WHEN eso.id IS NOT NULL AND eso.is_working_day = true THEN 'no_punch'
					ELSE 'day_off'
				END
			) AS status,
			COALESCE(da.is_late, false) AS is_late
		FROM generate_series($1::date, $2::date, '1 day') d
		CROSS JOIN employees e
		LEFT JOIN daily_attendances da ON da.employee_id = e.id AND da.date = d::date
		LEFT JOIN employee_work_patterns ewp ON ewp.employee_id = e.id
			AND ewp.valid_from <= d::date
			AND (ewp.valid_to IS NULL OR ewp.valid_to >= d::date)
			AND ewp.is_active = true
		LEFT JOIN work_patterns wp ON wp.id = ewp.work_pattern_id AND wp.is_active = true
		LEFT JOIN work_pattern_details wpd ON wpd.work_pattern_id = wp.id
			AND wpd.day_of_week = EXTRACT(DOW FROM d::date)
		LEFT JOIN employee_schedule_overrides eso ON eso.employee_id = e.id AND eso.date = d::date
		WHERE e.is_active = true
	)
`

func (r *PostgresDashboardRepo) CountMetrics(ctx context.Context, from, to time.Time) (*MetricsCounts, error) {
	var m MetricsCounts

	if err := r.db.GetContext(ctx, &m.TotalEmployees,
		`SELECT COUNT(*) FROM employees WHERE is_active = true`,
	); err != nil {
		return nil, fmt.Errorf("count employees: %w", err)
	}

	if err := r.db.GetContext(ctx, &m.ActiveContracts,
		`SELECT COUNT(*) FROM contracts WHERE status = 'active'`,
	); err != nil {
		return nil, fmt.Errorf("count active contracts: %w", err)
	}

	aggQuery := queryResolvedStatus + `
		SELECT
			COUNT(*) FILTER (WHERE status = 'present') AS present,
			COUNT(*) FILTER (WHERE status = 'absent') AS absent,
			COUNT(*) FILTER (WHERE status = 'on_leave') AS on_leave,
			COUNT(*) FILTER (WHERE is_late = true) AS late
		FROM resolved
	`
	if err := r.db.GetContext(ctx, &m, aggQuery, from, to); err != nil {
		return nil, fmt.Errorf("count attendance metrics: %w", err)
	}

	if err := r.db.GetContext(ctx, &m.PendingLeave,
		`SELECT COUNT(*) FROM leave_submissions WHERE status = 'pending'`,
	); err != nil {
		return nil, fmt.Errorf("count pending leave: %w", err)
	}

	return &m, nil
}

func (r *PostgresDashboardRepo) GetTrends(ctx context.Context, from, to time.Time) ([]TrendRow, error) {
	trendsQuery := queryResolvedStatus + `
		SELECT
			day::text AS date,
			COUNT(*) FILTER (WHERE status = 'present')::int AS present,
			COUNT(*) FILTER (WHERE status = 'on_leave')::int AS on_leave,
			COUNT(*) FILTER (WHERE status = 'absent')::int AS absent
		FROM resolved
		GROUP BY day
		ORDER BY day
	`
	var rows []TrendRow
	if err := r.db.SelectContext(ctx, &rows, trendsQuery, from, to); err != nil {
		return nil, fmt.Errorf("query trends: %w", err)
	}
	return rows, nil
}

var _ DashboardRepository = (*PostgresDashboardRepo)(nil)
