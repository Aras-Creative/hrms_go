package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AttendanceCorrection struct {
	ID               string
	EmployeeID       string
	Date             time.Time
	ClockIn          *time.Time
	ClockOut         *time.Time
	Status           *string
	IsLate           *bool
	IsEarlyLeave     *bool
	LeaveTypeName    *string
	LeaveSubmissionID *string
	Reason           string
	CorrectedBy      string
	CreatedAt        time.Time
	HasClockIn       bool
	HasClockOut      bool
}

func NewAttendanceCorrection(
	employeeID string,
	date time.Time,
	clockIn, clockOut *time.Time,
	status *string,
	isLate, isEarlyLeave *bool,
	leaveTypeName, leaveSubmissionID *string,
	reason, correctedBy string,
	hasClockIn, hasClockOut bool,
) *AttendanceCorrection {
	return &AttendanceCorrection{
		ID:               uuid.New().String(),
		EmployeeID:       employeeID,
		Date:             date,
		ClockIn:          clockIn,
		ClockOut:         clockOut,
		Status:           status,
		IsLate:           isLate,
		IsEarlyLeave:     isEarlyLeave,
		LeaveTypeName:    leaveTypeName,
		LeaveSubmissionID: leaveSubmissionID,
		Reason:           reason,
		CorrectedBy:      correctedBy,
		CreatedAt:        time.Now(),
		HasClockIn:       hasClockIn,
		HasClockOut:      hasClockOut,
	}
}

func ReconstituteAttendanceCorrection(
	id, employeeID string,
	date time.Time,
	clockIn, clockOut *time.Time,
	status *string,
	isLate, isEarlyLeave *bool,
	reason, correctedBy string,
	createdAt time.Time,
) *AttendanceCorrection {
	return &AttendanceCorrection{
		ID: id, EmployeeID: employeeID, Date: date,
		ClockIn: clockIn, ClockOut: clockOut, Status: status,
		IsLate: isLate, IsEarlyLeave: isEarlyLeave,
		Reason: reason, CorrectedBy: correctedBy, CreatedAt: createdAt,
	}
}

// Validate checks that the correction has valid business invariants.
func (c *AttendanceCorrection) Validate() error {
	if c.EmployeeID == "" {
		return fmt.Errorf("employee_id is required")
	}
	if c.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	hasField := c.HasClockIn || c.HasClockOut || c.Status != nil || c.IsLate != nil || c.IsEarlyLeave != nil || c.LeaveTypeName != nil || c.LeaveSubmissionID != nil
	if !hasField {
		return fmt.Errorf("at least one field to correct must be provided")
	}
	if c.ClockIn != nil && c.ClockOut != nil && c.ClockOut.Before(*c.ClockIn) {
		return fmt.Errorf("clock_out cannot be before clock_in")
	}
	if c.Status != nil {
		if _, err := ParseAttendanceStatus(*c.Status); err != nil {
			return err
		}
	}
	return nil
}

// ApplyTo applies this correction's values to a DailyAttendance record.
// It overwrites clock_in, clock_out, status, lateness flags,
// leave_type_name, leave_submission_id, total work seconds, and marks the source as "correction".
func (c *AttendanceCorrection) ApplyTo(da *DailyAttendance) {
	if c.HasClockIn {
		da.FirstPunchIn = c.ClockIn
	}
	if c.HasClockOut {
		da.LastPunchOut = c.ClockOut
	}
	if c.Status != nil {
		da.Status = AttendanceStatus(*c.Status)
	}
	if c.LeaveTypeName != nil {
		da.LeaveTypeName = c.LeaveTypeName
	}
	if c.LeaveSubmissionID != nil {
		da.LeaveSubmissionID = c.LeaveSubmissionID
	}
	if c.IsLate != nil {
		da.IsLate = *c.IsLate
	} else if c.LeaveSubmissionID != nil || da.LeaveSubmissionID != nil {
		da.IsLate = false
	} else {
		da.IsLate = da.LateMinutes() > 0
	}
	if c.IsEarlyLeave != nil {
		da.IsEarlyLeave = *c.IsEarlyLeave
	} else if c.LeaveSubmissionID != nil || da.LeaveSubmissionID != nil {
		da.IsEarlyLeave = false
	} else {
		da.IsEarlyLeave = da.EarlyLeaveMinutes() > 0
	}
	if da.FirstPunchIn != nil && da.LastPunchOut != nil {
		secs := int(da.LastPunchOut.Sub(*da.FirstPunchIn).Seconds())
		da.TotalWorkSeconds = &secs
	}
	da.Source = "correction"
	da.UpdatedAt = time.Now()
}
