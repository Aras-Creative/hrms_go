package delivery

import "time"

type CreateCorrectionRequest struct {
	EmployeeID       string     `json:"employee_id"`
	Date             string     `json:"date"`
	ClockIn          *time.Time `json:"clock_in"`
	ClockOut         *time.Time `json:"clock_out"`
	Status           *string    `json:"status"`
	IsLate           *bool      `json:"is_late"`
	IsEarlyLeave     *bool      `json:"is_early_leave"`
	LeaveTypeName    *string    `json:"leave_type_name"`
	LeaveSubmissionID *string   `json:"leave_submission_id"`
	Reason           string     `json:"reason"`
}
