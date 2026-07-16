package models

import "time"

type LeaveBalance struct {
	ID             string
	EmployeeID     string
	EmployeeName   string
	EmployeeNumber string
	ProfilePhotoID *string
	LeaveTypeID    string
	LeaveTypeName  string
	Year           int
	TotalDays      int
	UsedDays       int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateLeaveBalanceInput struct {
	EmployeeID  string
	LeaveTypeID string
	Year        int
	TotalDays   int
}

type ListBalanceInput struct {
	LeaveTypeID string
	Search      string
	Year        int
	Page        int
	PerPage     int
}

type ListBalanceResult struct {
	Items []LeaveBalance
	Total int64
}

type UpdateLeaveBalanceInput struct {
	EmployeeID  string
	LeaveTypeID string
	Year        int
	TotalDays   int
}
