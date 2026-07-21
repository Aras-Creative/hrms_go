package models

import (
	"time"
)

type PunchInput struct {
	EmployeeID string
}

type PunchEvent struct {
	EmployeeID       string     `json:"employee_id"`
	PunchType        string     `json:"punch_type"`
	Timestamp        time.Time  `json:"timestamp"`
	Status           string     `json:"status"`
	FirstPunchIn     *time.Time `json:"first_punch_in,omitempty"`
	LastPunchOut     *time.Time `json:"last_punch_out,omitempty"`
	LateMinutes      int        `json:"late_minutes"`
	TotalWorkSeconds *int       `json:"total_work_seconds,omitempty"`
}

type PunchHistoryInput struct {
	EmployeeID string
	From       string
	To         string
}

type MyAttendance struct {
	EmployeeID   string   `json:"employee_id"`
	EmployeeName string   `json:"employee_name"`
	Date         string   `json:"date"`
	ClockIn      *string  `json:"clock_in"`
	ClockOut     *string  `json:"clock_out"`
	TotalHours   float64  `json:"total_hours"`
	Status       string   `json:"status"`
	LateMinutes  int      `json:"late_minutes"`
	IsPresent    bool     `json:"is_present"`
	IsLate       bool     `json:"is_late"`
	IsAbsent     bool     `json:"is_absent"`
	IsComplete   bool     `json:"is_complete"`
}

type MyAttendanceHistoryItem struct {
	Day          string     `json:"day"`
	Month        string     `json:"month"`
	Type         string     `json:"type"`
	CheckIn      *string    `json:"check_in"`
	CheckOut     *string    `json:"check_out"`
	WorkingHours *string    `json:"working_hours"`
	Reason       *string    `json:"reason"`
	LateMinutes  int        `json:"late_minutes"`
	IsCorrected  bool       `json:"is_corrected"`
	CorrectedBy  *string    `json:"corrected_by,omitempty"`
	CorrectedAt  *time.Time `json:"corrected_at,omitempty"`
}

type MonthlyStats struct {
	Month       string `json:"month"`
	Present     int    `json:"present"`
	DayOff      int    `json:"day_off"`
	OnLeave     int    `json:"on_leave"`
	Absent      int    `json:"absent"`
	LateMinutes int    `json:"late_minutes"`
}

type ListInput struct {
	SearchName    string
	Status        string
	DesignationID string
	IsLate        string
	IsEarlyLeave  string
	From          string
	To            string
	Page          int
	PerPage       int
}

// AdminAttendanceItem is a read-model DTO for the admin daily attendance list.
type AdminAttendanceItem struct {
	ID                *string    `json:"id,omitempty"`
	EmployeeID        string     `json:"employee_id"`
	EmployeeName      string     `json:"employee_name"`
	EmployeeNumber    string     `json:"employee_number"`
	DesignationName   *string    `json:"designation_name,omitempty"`
	ProfilePhotoID    *string    `json:"profile_photo_id,omitempty"`
	Date              time.Time  `json:"date"`
	Status            string     `json:"status"`
	IsLate            bool       `json:"is_late"`
	IsEarlyLeave      bool       `json:"is_early_leave"`
	ExpectedStartTime *string    `json:"expected_start_time,omitempty"`
	ExpectedEndTime   *string    `json:"expected_end_time,omitempty"`
	Source            string     `json:"source"`
	FirstPunchIn      *time.Time `json:"first_punch_in,omitempty"`
	LastPunchOut      *time.Time `json:"last_punch_out,omitempty"`
	TotalWorkSeconds  *int       `json:"total_work_seconds,omitempty"`
	LeaveSubmissionID *string    `json:"leave_submission_id,omitempty"`
	LeaveTypeName     *string    `json:"leave_type_name,omitempty"`
	ScheduleOverrideID *string   `json:"schedule_override_id,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}

// CorrectionViewItem is a read-model DTO for the correction list view.
type CorrectionViewItem struct {
	ID               string     `json:"id"`
	EmployeeID       string     `json:"employee_id"`
	EmployeeName     string     `json:"employee_name"`
	Date             time.Time  `json:"date"`
	ClockIn          *time.Time `json:"clock_in,omitempty"`
	ClockOut         *time.Time `json:"clock_out,omitempty"`
	Status           *string    `json:"status,omitempty"`
	IsLate           *bool      `json:"is_late,omitempty"`
	IsEarlyLeave     *bool      `json:"is_early_leave,omitempty"`
	LeaveTypeName    *string    `json:"leave_type_name,omitempty"`
	LeaveSubmissionID *string   `json:"leave_submission_id,omitempty"`
	Reason           string     `json:"reason"`
	CorrectedBy      string     `json:"corrected_by"`
	CreatedAt        time.Time  `json:"created_at"`
}

type ListResult struct {
	Items []*AdminAttendanceItem
	Total int64
}
