package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LeaveBalance struct {
	ID          string
	EmployeeID  string
	LeaveTypeID string
	Year        int
	TotalDays   int
	UsedDays    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewLeaveBalance(employeeID, leaveTypeID string, year, totalDays int) *LeaveBalance {
	now := time.Now()
	return &LeaveBalance{
		ID:          uuid.New().String(),
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
		TotalDays:   totalDays,
		UsedDays:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func ReconstituteLeaveBalance(id, employeeID, leaveTypeID string, year, totalDays, usedDays int, createdAt, updatedAt time.Time) *LeaveBalance {
	return &LeaveBalance{
		ID:          id,
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
		TotalDays:   totalDays,
		UsedDays:    usedDays,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func (lb *LeaveBalance) SetTotalDays(days int) {
	lb.TotalDays = days
	lb.UpdatedAt = time.Now()
}

func (lb *LeaveBalance) SetUsedDays(days int) {
	lb.UsedDays = days
	if lb.UsedDays < 0 {
		lb.UsedDays = 0
	}
	lb.UpdatedAt = time.Now()
}

func (lb *LeaveBalance) Restore(days int) {
	lb.UsedDays -= days
	if lb.UsedDays < 0 {
		lb.UsedDays = 0
	}
	lb.UpdatedAt = time.Now()
}

func (lb *LeaveBalance) Consume(days int) error {
	if days <= 0 {
		return fmt.Errorf("consumed days must be positive")
	}
	if lb.UsedDays+days > lb.TotalDays {
		return fmt.Errorf("insufficient leave balance: %d remaining, %d requested", lb.Remaining(), days)
	}
	lb.UsedDays += days
	lb.UpdatedAt = time.Now()
	return nil
}

func (lb *LeaveBalance) SufficientQuota(days int) bool {
	return days > 0 && lb.UsedDays+days <= lb.TotalDays
}

func (lb *LeaveBalance) Remaining() int {
	return lb.TotalDays - lb.UsedDays
}
