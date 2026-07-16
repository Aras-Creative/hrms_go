package models

import (
	"time"
)

type LeaveSubmission struct {
	ID             string
	EmployeeID     string
	EmployeeName   string
	EmployeeNumber string
	ProfilePhotoID *string
	LeaveTypeID    string
	LeaveTypeName  string
	StartDate      time.Time
	EndDate        time.Time
	Days           int
	Reason         string
	AttachmentID   *string
	Status         string
	ApprovedBy     *string
	ApprovedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateLeaveSubmissionInput struct {
	UserID       string
	LeaveTypeID  string
	StartDate    time.Time
	EndDate      time.Time
	Reason       string
	AttachmentID *string
}

type ListSubmissionInput struct {
	UserID  string
	Status  string
	Page    int
	PerPage int
}

type ListAllSubmissionInput struct {
	Status    string
	Search    string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PerPage   int
}

type MySubmissionResult struct {
	Items []LeaveSubmission
	Total int64
}

type ListSubmissionResult struct {
	Items []LeaveSubmission
	Total int64
}
