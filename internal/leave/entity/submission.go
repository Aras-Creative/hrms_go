package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CountWeekdays returns the number of weekdays (Mon-Fri) between start and end inclusive.
func CountWeekdays(start, end time.Time) int {
	days := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if w := d.Weekday(); w != time.Saturday && w != time.Sunday {
			days++
		}
	}
	return days
}

type LeaveSubmission struct {
	ID           string
	EmployeeID   string
	LeaveTypeID  string
	StartDate    time.Time
	EndDate      time.Time
	Days         int
	Reason       string
	AttachmentID *string
	Status       LeaveStatus
	ApprovedBy   *string
	ApprovedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewLeaveSubmission(employeeID, leaveTypeID string, startDate, endDate time.Time, days int, reason string, attachmentID *string) *LeaveSubmission {
	now := time.Now()
	return &LeaveSubmission{
		ID:           uuid.New().String(),
		EmployeeID:   employeeID,
		LeaveTypeID:  leaveTypeID,
		StartDate:    startDate,
		EndDate:      endDate,
		Days:         days,
		Reason:       reason,
		AttachmentID: attachmentID,
		Status:       LeaveStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func ReconstituteLeaveSubmission(
	id, employeeID, leaveTypeID string,
	startDate, endDate time.Time,
	days int,
	reason string,
	attachmentID *string,
	status LeaveStatus,
	approvedBy *string,
	approvedAt *time.Time,
	createdAt, updatedAt time.Time,
) *LeaveSubmission {
	return &LeaveSubmission{
		ID:           id,
		EmployeeID:   employeeID,
		LeaveTypeID:  leaveTypeID,
		StartDate:    startDate,
		EndDate:      endDate,
		Days:         days,
		Reason:       reason,
		AttachmentID: attachmentID,
		Status:       status,
		ApprovedBy:   approvedBy,
		ApprovedAt:   approvedAt,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func (s *LeaveSubmission) Approve(approvedBy string) error {
	if !s.Status.CanTransitionTo(LeaveStatusApproved) {
		return fmt.Errorf("cannot approve submission with status %s", s.Status)
	}
	now := time.Now()
	s.Status = LeaveStatusApproved
	s.ApprovedBy = &approvedBy
	s.ApprovedAt = &now
	s.UpdatedAt = now
	return nil
}

func (s *LeaveSubmission) Reject() error {
	if !s.Status.CanTransitionTo(LeaveStatusRejected) {
		return fmt.Errorf("cannot reject submission with status %s", s.Status)
	}
	s.Status = LeaveStatusRejected
	s.UpdatedAt = time.Now()
	return nil
}

func (s *LeaveSubmission) Cancel() error {
	if !s.Status.CanTransitionTo(LeaveStatusCancelled) {
		return fmt.Errorf("cannot cancel submission with status %s", s.Status)
	}
	s.Status = LeaveStatusCancelled
	s.UpdatedAt = time.Now()
	return nil
}
