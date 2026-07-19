package repository

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/pkg/timeutil"
)

type RecapRow struct {
	EmployeeID       string  `db:"employee_id"`
	EmployeeNumber   string  `db:"employee_number"`
	FullName         string  `db:"full_name"`
	ProfilePhotoID   *string `db:"profile_photo_id"`
	DesignationName  *string `db:"designation_name"`
	WorkingDays      int     `db:"working_days"`
	Present          int     `db:"present"`
	LeaveTypeName    *string `db:"leave_type_name"`
	Absent           int     `db:"absent"`
	MissingClockOut  int     `db:"missing_clock_out"`
	Late             int     `db:"late"`
	LateMinutes      int     `db:"late_minutes"`
	EarlyLeave       int     `db:"early_leave"`
}

type LeaveTypeRow struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

type workingDaysRow struct {
	EmployeeID  string `db:"employee_id"`
	WorkingDays int    `db:"working_days"`
}

const queryActiveLeaveTypes = `SELECT id, name FROM leave_types WHERE is_active = true ORDER BY name ASC`

const queryWorkingDaysBase = `
	WITH dow_counts AS (
		SELECT (EXTRACT(DOW FROM d))::int AS dow, COUNT(*) AS cnt
		FROM generate_series($1::date, $2::date, '1 day') d
		GROUP BY 1
	),
	pattern_days AS (
		SELECT
			ewp.employee_id,
			SUM(dc.cnt) AS days
		FROM employee_work_patterns ewp
		JOIN work_patterns wp ON wp.id = ewp.work_pattern_id AND wp.is_active = true
		JOIN work_pattern_details wpd ON wpd.work_pattern_id = wp.id
			AND (wpd.working_type IS NULL OR wpd.working_type != 'off')
			AND (wpd.working_type = 'dynamic' OR wpd.start_time IS NOT NULL)
		JOIN dow_counts dc ON dc.dow = wpd.day_of_week
		WHERE ewp.valid_from <= $2
			AND (ewp.valid_to IS NULL OR ewp.valid_to >= $1)
			AND ewp.is_active = true
		GROUP BY ewp.employee_id
	),
	override_adj AS (
		SELECT
			o.employee_id,
			-- force_work: override says working on a date whose DOW is NOT a working day in the pattern
			COUNT(*) FILTER (
				WHERE o.is_working_day = true
				AND NOT EXISTS (
					SELECT 1
					FROM employee_work_patterns ewp2
					JOIN work_patterns wp2 ON wp2.id = ewp2.work_pattern_id AND wp2.is_active = true
					JOIN work_pattern_details wpd2 ON wpd2.work_pattern_id = wp2.id
						AND (wpd2.working_type IS NULL OR wpd2.working_type != 'off')
						AND (wpd2.working_type = 'dynamic' OR wpd2.start_time IS NOT NULL)
					WHERE ewp2.employee_id = o.employee_id
						AND wpd2.day_of_week = EXTRACT(DOW FROM o.date)
						AND ewp2.valid_from <= o.date
						AND (ewp2.valid_to IS NULL OR ewp2.valid_to >= o.date)
						AND ewp2.is_active = true
				)
			) AS force_work,
			-- force_off: override says not working on a date whose DOW IS a working day in the pattern
			COUNT(*) FILTER (
				WHERE o.is_working_day = false
				AND EXISTS (
					SELECT 1
					FROM employee_work_patterns ewp3
					JOIN work_patterns wp3 ON wp3.id = ewp3.work_pattern_id AND wp3.is_active = true
					JOIN work_pattern_details wpd3 ON wpd3.work_pattern_id = wp3.id
						AND (wpd3.working_type IS NULL OR wpd3.working_type != 'off')
						AND (wpd3.working_type = 'dynamic' OR wpd3.start_time IS NOT NULL)
					WHERE ewp3.employee_id = o.employee_id
						AND wpd3.day_of_week = EXTRACT(DOW FROM o.date)
						AND ewp3.valid_from <= o.date
						AND (ewp3.valid_to IS NULL OR ewp3.valid_to >= o.date)
						AND ewp3.is_active = true
				)
			) AS force_off
		FROM employee_schedule_overrides o
		WHERE o.date >= $1 AND o.date <= $2
		GROUP BY o.employee_id
	)
	SELECT
		e.id AS employee_id,
		GREATEST(0,
			COALESCE(pd.days, 0)
			+ COALESCE(oa.force_work, 0)
			- COALESCE(oa.force_off, 0)
		) AS working_days
	FROM employees e
	LEFT JOIN pattern_days pd ON pd.employee_id = e.id
	LEFT JOIN override_adj oa ON oa.employee_id = e.id
	WHERE 1=1
`

const queryAttendanceMetricsBase = `
	SELECT
		e.id AS employee_id,
		e.employee_number,
		e.full_name,
		e.profile_photo_id,
		des.name AS designation_name,
		COUNT(da.id) FILTER (WHERE da.status = 'present') AS present,
		lt.name AS leave_type_name,
		COUNT(da.id) FILTER (WHERE da.status = 'absent') AS absent,
		COUNT(da.id) FILTER (WHERE da.first_punch_in IS NOT NULL AND da.last_punch_out IS NULL) AS missing_clock_out,
		COUNT(da.id) FILTER (WHERE da.is_late = true) AS late,
		COUNT(da.id) FILTER (WHERE da.is_early_leave = true) AS early_leave,
		COALESCE(SUM(
			CASE WHEN da.is_late = true AND da.first_punch_in IS NOT NULL AND da.expected_start_time IS NOT NULL THEN
				GREATEST(0, EXTRACT(EPOCH FROM (
					da.first_punch_in - ((da.date + da.expected_start_time::time) AT TIME ZONE %s)
				))::int / 60)
			ELSE 0 END
		), 0) AS late_minutes
	FROM employees e
	LEFT JOIN daily_attendances da ON da.employee_id = e.id AND da.date >= $1 AND da.date <= $2
	LEFT JOIN designations des ON des.id = e.designation_id
	LEFT JOIN leave_submissions ls ON ls.id = da.leave_submission_id
	LEFT JOIN leave_types lt ON lt.id = ls.leave_type_id
	WHERE 1=1
`

func (r *PostgresDailyAttendanceRepo) Recap(ctx context.Context, from, to time.Time, designationID string) ([]*RecapRow, error) {
	args := []interface{}{from, to}
	argIdx := 3
	where := ""

	if designationID != "" {
		where += fmt.Sprintf(" AND e.designation_id = $%d::uuid", argIdx)
		args = append(args, designationID)
		argIdx++
	}

	tz := fmt.Sprintf("'%s'", timeutil.DefaultTimezone)
	metricsQuery := fmt.Sprintf(queryAttendanceMetricsBase, tz) + where + `
		GROUP BY e.id, e.employee_number, e.full_name, e.profile_photo_id, des.name, lt.name
		ORDER BY e.full_name ASC, lt.name ASC NULLS LAST
	`
	var metrics []*RecapRow
	if err := r.db.SelectContext(ctx, &metrics, metricsQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to get attendance metrics: %w", err)
	}

	workingDaysQuery := queryWorkingDaysBase + where
	var workDays []workingDaysRow
	if err := r.db.SelectContext(ctx, &workDays, workingDaysQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to get working days: %w", err)
	}

	wdMap := make(map[string]int, len(workDays))
	for _, w := range workDays {
		wdMap[w.EmployeeID] = w.WorkingDays
	}

	for _, m := range metrics {
		m.WorkingDays = wdMap[m.EmployeeID]
	}

	return metrics, nil
}

func (r *PostgresDailyAttendanceRepo) FindActiveLeaveTypes(ctx context.Context) ([]LeaveTypeRow, error) {
	var list []LeaveTypeRow
	if err := r.db.SelectContext(ctx, &list, queryActiveLeaveTypes); err != nil {
		return nil, fmt.Errorf("failed to find active leave types: %w", err)
	}
	return list, nil
}
