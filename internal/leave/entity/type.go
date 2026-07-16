package entity

import (
	"time"

	"github.com/google/uuid"
)

type LeaveType struct {
	ID          string
	Name        string
	DefaultDays int
	IsPaid      bool
	IsUnlimited bool
	IsHalfDay   bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewLeaveType(name string, defaultDays int, isPaid, isUnlimited, isHalfDay bool) *LeaveType {
	now := time.Now()
	return &LeaveType{
		ID:          uuid.New().String(),
		Name:        name,
		DefaultDays: defaultDays,
		IsPaid:      isPaid,
		IsUnlimited: isUnlimited,
		IsHalfDay:   isHalfDay,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func ReconstituteLeaveType(id, name string, defaultDays int, isPaid, isUnlimited, isHalfDay, isActive bool, createdAt, updatedAt time.Time) *LeaveType {
	return &LeaveType{
		ID:          id,
		Name:        name,
		DefaultDays: defaultDays,
		IsPaid:      isPaid,
		IsUnlimited: isUnlimited,
		IsHalfDay:   isHalfDay,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func (lt *LeaveType) Rename(name string) {
	lt.Name = name
	lt.UpdatedAt = time.Now()
}

func (lt *LeaveType) SetDefaultDays(days int) {
	lt.DefaultDays = days
	lt.UpdatedAt = time.Now()
}

func (lt *LeaveType) SetPaidStatus(paid bool) {
	lt.IsPaid = paid
	lt.UpdatedAt = time.Now()
}

func (lt *LeaveType) SetUnlimited(unlimited bool) {
	lt.IsUnlimited = unlimited
	lt.UpdatedAt = time.Now()
}

func (lt *LeaveType) Disable() {
	lt.IsActive = false
	lt.UpdatedAt = time.Now()
}

func (lt *LeaveType) Enable() {
	lt.IsActive = true
	lt.UpdatedAt = time.Now()
}
