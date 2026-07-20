package repository

import "time"

type DailyAttendanceModel struct {
	ID                 string     `db:"id"`
	EmployeeID         string     `db:"employee_id"`
	Date               time.Time  `db:"date"`
	Status             string     `db:"status"`
	IsLate             bool       `db:"is_late"`
	IsEarlyLeave       bool       `db:"is_early_leave"`
	ExpectedStartTime  *string    `db:"expected_start_time"`
	ExpectedEndTime    *string    `db:"expected_end_time"`
	Source             string     `db:"source"`
	FirstPunchIn       *time.Time `db:"first_punch_in"`
	LastPunchOut       *time.Time `db:"last_punch_out"`
	TotalWorkSeconds   *int       `db:"total_work_seconds"`
	LeaveSubmissionID  *string    `db:"leave_submission_id"`
	LeaveTypeName      *string    `db:"leave_type_name"`
	ScheduleOverrideID *string    `db:"schedule_override_id"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

type DailyComputationRow struct {
	EmployeeID         string     `db:"employee_id"`
	Date               time.Time  `db:"date"`
	ExpectedStartTime  *string    `db:"expected_start_time"`
	ExpectedEndTime    *string    `db:"expected_end_time"`
	Source             string     `db:"source"`
	FirstPunchIn       *time.Time `db:"first_punch_in"`
	LastPunchOut       *time.Time `db:"last_punch_out"`
	TotalWorkSeconds   *int       `db:"total_work_seconds"`
	LeaveSubmissionID  *string    `db:"leave_submission_id"`
	LeaveTypeName      *string    `db:"leave_type_name"`
	ScheduleOverrideID *string    `db:"schedule_override_id"`
	OverrideIsWorking  *bool      `db:"override_is_working"`
	LeaveIsHalfDay     *bool      `db:"leave_is_half_day"`
	WorkingType        string     `db:"-"`
}

type CorrectionModel struct {
	ID           string     `db:"id"`
	EmployeeID   string     `db:"employee_id"`
	Date         time.Time  `db:"date"`
	ClockIn      *time.Time `db:"clock_in"`
	ClockOut     *time.Time `db:"clock_out"`
	Status       *string    `db:"status"`
	IsLate       *bool      `db:"is_late"`
	IsEarlyLeave *bool      `db:"is_early_leave"`
	Reason       string     `db:"reason"`
	CorrectedBy  string     `db:"corrected_by"`
	CreatedAt    time.Time  `db:"created_at"`
}

type PunchModel struct {
	ID         string    `db:"id"`
	EmployeeID string    `db:"employee_id"`
	Type       string    `db:"type"`
	Timestamp  time.Time `db:"timestamp"`
	Date       time.Time `db:"date"`
	CreatedAt  time.Time `db:"created_at"`
}

type AdminAttendanceRow struct {
	ID                 *string    `db:"id"`
	EmployeeID         string     `db:"employee_id"`
	Date               time.Time  `db:"date"`
	Status             string     `db:"status"`
	IsLate             bool       `db:"is_late"`
	IsEarlyLeave       bool       `db:"is_early_leave"`
	ExpectedStartTime  *string    `db:"expected_start_time"`
	ExpectedEndTime    *string    `db:"expected_end_time"`
	Source             string     `db:"source"`
	FirstPunchIn       *time.Time `db:"first_punch_in"`
	LastPunchOut       *time.Time `db:"last_punch_out"`
	TotalWorkSeconds   *int       `db:"total_work_seconds"`
	LeaveSubmissionID  *string    `db:"leave_submission_id"`
	LeaveTypeName      *string    `db:"leave_type_name"`
	ScheduleOverrideID *string    `db:"schedule_override_id"`
	CreatedAt          *time.Time `db:"created_at"`
	UpdatedAt          *time.Time `db:"updated_at"`
	EmployeeName       string     `db:"employee_name"`
	EmployeeNumber     string     `db:"employee_number"`
	DesignationName    *string    `db:"designation_name"`
	ProfilePhotoID     *string    `db:"profile_photo_id"`
	OverrideIsWorking  *bool      `db:"override_is_working"`
	OverrideStartTime  *string    `db:"override_start_time"`
	OverrideEndTime    *string    `db:"override_end_time"`
	PatternStartTime   *string    `db:"pattern_start_time"`
	PatternEndTime     *string    `db:"pattern_end_time"`
	PatternType        *string    `db:"pattern_working_type"`
}
