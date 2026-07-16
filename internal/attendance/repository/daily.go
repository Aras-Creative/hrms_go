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
	queryUpsertDailyAttendance = `
		INSERT INTO daily_attendances (
			id, employee_id, date, status,
			is_late, is_early_leave,
			expected_start_time, expected_end_time, source,
			first_punch_in, last_punch_out, total_work_seconds,
			leave_submission_id, leave_type_name, schedule_override_id,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (employee_id, date) DO UPDATE SET
			status = EXCLUDED.status,
			is_late = EXCLUDED.is_late,
			is_early_leave = EXCLUDED.is_early_leave,
			expected_start_time = EXCLUDED.expected_start_time,
			expected_end_time = EXCLUDED.expected_end_time,
			source = EXCLUDED.source,
			first_punch_in = EXCLUDED.first_punch_in,
			last_punch_out = EXCLUDED.last_punch_out,
			total_work_seconds = EXCLUDED.total_work_seconds,
			leave_submission_id = EXCLUDED.leave_submission_id,
			leave_type_name = EXCLUDED.leave_type_name,
			schedule_override_id = EXCLUDED.schedule_override_id,
			updated_at = EXCLUDED.updated_at
	`

	querySelectDailyAttendance = `
		SELECT id, employee_id, date, status,
			is_late, is_early_leave,
			expected_start_time, expected_end_time, source,
			first_punch_in, last_punch_out, total_work_seconds,
			leave_submission_id, leave_type_name, schedule_override_id,
			created_at, updated_at
		FROM daily_attendances
	`
)

var (
	queryDailyByID                   = querySelectDailyAttendance + ` WHERE id = $1`
	queryDailyByEmployeeRange        = querySelectDailyAttendance + ` WHERE employee_id = $1 AND date >= $2 AND date <= $3 ORDER BY date ASC`
	queryDailyByEmployeeAndDate      = querySelectDailyAttendance + ` WHERE employee_id = $1 AND date = $2`
)

type PostgresDailyAttendanceRepo struct {
	db txContext
}

func NewPostgresDailyAttendanceRepo(db *sqlx.DB) *PostgresDailyAttendanceRepo {
	return &PostgresDailyAttendanceRepo{db: db}
}

func (r *PostgresDailyAttendanceRepo) WithTx(tx *sqlx.Tx) DailyAttendanceRepository {
	return &PostgresDailyAttendanceRepo{db: tx}
}

func (r *PostgresDailyAttendanceRepo) Upsert(ctx context.Context, da *entity.DailyAttendance) error {
	_, err := r.db.ExecContext(ctx, queryUpsertDailyAttendance,
		da.ID, da.EmployeeID, da.Date, string(da.Status),
		da.IsLate, da.IsEarlyLeave,
		da.ExpectedStartTime, da.ExpectedEndTime, da.Source,
		da.FirstPunchIn, da.LastPunchOut, da.TotalWorkSeconds,
		da.LeaveSubmissionID, da.LeaveTypeName, da.ScheduleOverrideID,
		da.CreatedAt, da.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert daily attendance: %w", err)
	}
	return nil
}

func (r *PostgresDailyAttendanceRepo) FindByID(ctx context.Context, id string) (*entity.DailyAttendance, error) {
	var m DailyAttendanceModel
	if err := r.db.GetContext(ctx, &m, queryDailyByID, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find daily attendance by id: %w", err)
	}
	return modelToDailyAttendance(&m), nil
}

func (r *PostgresDailyAttendanceRepo) FindByEmployeeAndDateRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.DailyAttendance, error) {
	var models []DailyAttendanceModel
	if err := r.db.SelectContext(ctx, &models, queryDailyByEmployeeRange, employeeID, from, to); err != nil {
		return nil, fmt.Errorf("failed to find daily attendances: %w", err)
	}
	result := make([]*entity.DailyAttendance, 0, len(models))
	for _, m := range models {
		result = append(result, modelToDailyAttendance(&m))
	}
	return result, nil
}

func (r *PostgresDailyAttendanceRepo) FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*entity.DailyAttendance, error) {
	var m DailyAttendanceModel
	if err := r.db.GetContext(ctx, &m, queryDailyByEmployeeAndDate, employeeID, date); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find daily attendance: %w", err)
	}
	return modelToDailyAttendance(&m), nil
}

func modelToDailyAttendance(m *DailyAttendanceModel) *entity.DailyAttendance {
	return entity.ReconstituteDailyAttendance(
		m.ID, m.EmployeeID, m.Date,
		entity.AttendanceStatus(m.Status),
		m.IsLate, m.IsEarlyLeave,
		m.ExpectedStartTime, m.ExpectedEndTime, m.Source,
		m.FirstPunchIn, m.LastPunchOut, m.TotalWorkSeconds,
		m.LeaveSubmissionID, m.LeaveTypeName, m.ScheduleOverrideID,
		m.CreatedAt, m.UpdatedAt,
	)
}
