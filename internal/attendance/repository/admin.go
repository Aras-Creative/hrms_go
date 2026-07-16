package repository

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/attendance/entity"
)

const querySelectAdminAttendance = `
	SELECT
		da.id,
		e.id AS employee_id,
		date_series.d::date AS date,
		COALESCE(da.status, '') AS status,
		COALESCE(da.is_late, false) AS is_late,
		COALESCE(da.is_early_leave, false) AS is_early_leave,
		da.expected_start_time,
		da.expected_end_time,
		COALESCE(da.source, '') AS source,
		da.first_punch_in,
		da.last_punch_out,
		da.total_work_seconds,
		da.leave_submission_id,
		da.leave_type_name,
		da.schedule_override_id,
		da.created_at,
		da.updated_at,
		e.full_name AS employee_name,
		e.employee_number,
		e.profile_photo_id,
		des.name AS designation_name,
		eso.is_working_day AS override_is_working,
		eso.start_time AS override_start_time,
		eso.end_time AS override_end_time,
		wpd.start_time AS pattern_start_time,
		wpd.end_time AS pattern_end_time,
		wpd.working_type AS pattern_working_type
	FROM generate_series($1::date, $2::date, '1 day') date_series(d)
	CROSS JOIN employees e
	LEFT JOIN designations des ON des.id = e.designation_id
	LEFT JOIN daily_attendances da ON da.employee_id = e.id AND da.date = date_series.d::date
	LEFT JOIN employee_work_patterns ewp ON ewp.employee_id = e.id
		AND ewp.valid_from <= date_series.d::date
		AND (ewp.valid_to IS NULL OR ewp.valid_to >= date_series.d::date)
		AND ewp.is_active = true
	LEFT JOIN work_patterns wp ON wp.id = ewp.work_pattern_id AND wp.is_active = true
	LEFT JOIN work_pattern_details wpd ON wpd.work_pattern_id = wp.id
		AND wpd.day_of_week = EXTRACT(DOW FROM date_series.d)
	LEFT JOIN employee_schedule_overrides eso ON eso.employee_id = e.id
		AND eso.date = date_series.d::date
`

func (r *PostgresDailyAttendanceRepo) FindAllPaginated(ctx context.Context, searchName, status, designationID, isLate, isEarlyLeave string, from, to time.Time, page, perPage int) ([]*AdminAttendanceRow, int64, error) {
	where := " WHERE 1=1"
	args := []interface{}{from, to}
	argIdx := 3

	if searchName != "" {
		where += fmt.Sprintf(" AND e.full_name ILIKE $%d", argIdx)
		args = append(args, "%"+searchName+"%")
		argIdx++
	}
	if designationID != "" {
		where += fmt.Sprintf(" AND e.designation_id = $%d", argIdx)
		args = append(args, designationID)
		argIdx++
	}
	if status != "" {
		where += fmt.Sprintf(` AND COALESCE(da.status,
			CASE
				WHEN eso.id IS NOT NULL AND eso.is_working_day = false THEN 'day_off'
				WHEN wpd.id IS NOT NULL AND wpd.working_type = 'dynamic' THEN 'no_punch'
				WHEN wpd.id IS NOT NULL AND wpd.start_time IS NOT NULL AND (wpd.working_type IS NULL OR wpd.working_type != 'off') THEN 'no_punch'
				WHEN eso.id IS NOT NULL AND eso.is_working_day = true THEN 'no_punch'
				ELSE 'day_off'
			END
		) = $%d`, argIdx)
		args = append(args, status)
		argIdx++
	}
	if isLate != "" {
		if isLate == "true" {
			where += " AND da.is_late = true"
		} else if isLate == "false" {
			where += " AND (da.is_late = false OR da.is_late IS NULL)"
		}
	}
	if isEarlyLeave != "" {
		if isEarlyLeave == "true" {
			where += " AND da.is_early_leave = true"
		} else if isEarlyLeave == "false" {
			where += " AND (da.is_early_leave = false OR da.is_early_leave IS NULL)"
		}
	}

	countQuery := `SELECT COUNT(*) FROM generate_series($1::date, $2::date, '1 day') date_series(d)
		CROSS JOIN employees e
		LEFT JOIN designations des ON des.id = e.designation_id
		LEFT JOIN daily_attendances da ON da.employee_id = e.id AND da.date = date_series.d::date
		LEFT JOIN employee_work_patterns ewp ON ewp.employee_id = e.id
			AND ewp.valid_from <= date_series.d::date
			AND (ewp.valid_to IS NULL OR ewp.valid_to >= date_series.d::date)
			AND ewp.is_active = true
		LEFT JOIN work_patterns wp ON wp.id = ewp.work_pattern_id AND wp.is_active = true
		LEFT JOIN work_pattern_details wpd ON wpd.work_pattern_id = wp.id
			AND wpd.day_of_week = EXTRACT(DOW FROM date_series.d)
		LEFT JOIN employee_schedule_overrides eso ON eso.employee_id = e.id
			AND eso.date = date_series.d::date` + where
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count admin attendance: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := querySelectAdminAttendance + where + fmt.Sprintf(" ORDER BY e.full_name ASC, date_series.d ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list admin attendance: %w", err)
	}
	defer rows.Close()

	list := make([]*AdminAttendanceRow, 0)
	for rows.Next() {
		var m AdminAttendanceRow
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("failed to scan admin attendance row: %w", err)
		}
		sf := entity.AdminScheduleFields{
			Status:             m.Status,
			Source:             m.Source,
			Date:               m.Date,
			ExpectedStartTime:  m.ExpectedStartTime,
			ExpectedEndTime:    m.ExpectedEndTime,
			ScheduleOverrideID: m.ScheduleOverrideID,
			OverrideIsWorking:  m.OverrideIsWorking,
			OverrideStartTime:  m.OverrideStartTime,
			OverrideEndTime:    m.OverrideEndTime,
			PatternStartTime:   m.PatternStartTime,
			PatternEndTime:     m.PatternEndTime,
			PatternType:        m.PatternType,
		}
		entity.ResolveAdminAttendance(&sf)
		m.Status = sf.Status
		m.Source = sf.Source
		m.ExpectedStartTime = sf.ExpectedStartTime
		m.ExpectedEndTime = sf.ExpectedEndTime
		m.ScheduleOverrideID = sf.ScheduleOverrideID
		list = append(list, &m)
	}
	return list, total, rows.Err()
}

var _ DailyAttendanceRepository = (*PostgresDailyAttendanceRepo)(nil)
