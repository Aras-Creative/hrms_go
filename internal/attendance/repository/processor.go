package repository

import (
	"context"
	"fmt"
	"time"
)

const (
	queryComputeDailyAttendance = `
		SELECT
			e.id AS employee_id,
			$2::date AS date,
			punch_in.ts AS first_punch_in,
			punch_out.ts AS last_punch_out,
			EXTRACT(EPOCH FROM (punch_out.ts - punch_in.ts))::int AS total_work_seconds,
			ls.id AS leave_submission_id,
			lt.name AS leave_type_name,
			lt.is_half_day AS leave_is_half_day
		FROM employees e
		LEFT JOIN LATERAL (
			SELECT timestamp AS ts FROM punches
			WHERE employee_id = e.id AND type = 'in' AND punches.date = $2::date
			ORDER BY timestamp ASC LIMIT 1
		) punch_in ON true
		LEFT JOIN LATERAL (
			SELECT timestamp AS ts FROM punches
			WHERE employee_id = e.id AND type = 'out' AND punches.date = $2::date
			ORDER BY timestamp DESC LIMIT 1
		) punch_out ON true
		LEFT JOIN leave_submissions ls ON ls.employee_id = e.id
			AND ls.status = 'approved'
			AND ls.start_date <= $2::date
			AND ls.end_date >= $2::date
		LEFT JOIN leave_types lt ON lt.id = ls.leave_type_id
		WHERE e.id = $1
	`

	queryComputeRangeAttendance = `
		SELECT
			e.id AS employee_id,
			date_series.d::date AS date,
			punch_in.ts AS first_punch_in,
			punch_out.ts AS last_punch_out,
			EXTRACT(EPOCH FROM (punch_out.ts - punch_in.ts))::int AS total_work_seconds,
			ls.id AS leave_submission_id,
			lt.name AS leave_type_name,
			lt.is_half_day AS leave_is_half_day
		FROM generate_series($1::date, $2::date, '1 day') date_series(d)
		CROSS JOIN employees e
		LEFT JOIN LATERAL (
			SELECT timestamp AS ts FROM punches
			WHERE employee_id = e.id AND type = 'in' AND punches.date = date_series.d::date
			ORDER BY timestamp ASC LIMIT 1
		) punch_in ON true
		LEFT JOIN LATERAL (
			SELECT timestamp AS ts FROM punches
			WHERE employee_id = e.id AND type = 'out' AND punches.date = date_series.d::date
			ORDER BY timestamp DESC LIMIT 1
		) punch_out ON true
		LEFT JOIN leave_submissions ls ON ls.employee_id = e.id
			AND ls.status = 'approved'
			AND ls.start_date <= date_series.d::date
			AND ls.end_date >= date_series.d::date
		LEFT JOIN leave_types lt ON lt.id = ls.leave_type_id
		WHERE NOT EXISTS (
			SELECT 1 FROM attendance_corrections ac
			WHERE ac.employee_id = e.id AND ac.date = date_series.d::date
		)
		ORDER BY e.id, date_series.d
	`
)

func (r *PostgresDailyAttendanceRepo) ComputeDaily(ctx context.Context, employeeID string, date time.Time) (*DailyComputationRow, error) {
	var row DailyComputationRow
	if err := r.db.GetContext(ctx, &row, queryComputeDailyAttendance, employeeID, date); err != nil {
		return nil, fmt.Errorf("failed to compute daily attendance: %w", err)
	}
	return &row, nil
}

func (r *PostgresDailyAttendanceRepo) ComputeRange(ctx context.Context, from, to time.Time) ([]DailyComputationRow, error) {
	var rows []DailyComputationRow
	if err := r.db.SelectContext(ctx, &rows, queryComputeRangeAttendance, from, to); err != nil {
		return nil, fmt.Errorf("failed to compute range attendance: %w", err)
	}
	return rows, nil
}
