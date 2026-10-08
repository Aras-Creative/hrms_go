package models

import "time"

type LeaveType struct {
	ID                           string
	Name                         string
	DefaultDays                  int
	IsPaid                       bool
	IsUnlimited                  bool
	IsHalfDay                    bool
	IncludeInAttendanceAllowance bool
	IsActive                     bool
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type CreateLeaveTypeInput struct {
	Name                         string
	DefaultDays                  int
	IsPaid                       bool
	IsUnlimited                  bool
	IsHalfDay                    bool
	IncludeInAttendanceAllowance bool
}

type UpdateLeaveTypeInput struct {
	Name                         string
	DefaultDays                  int
	IsPaid                       bool
	IsUnlimited                  bool
	IsHalfDay                    bool
	IncludeInAttendanceAllowance bool
}
