package entity

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeScheduleOverride struct {
	ID           string
	EmployeeID   string
	Date         time.Time
	IsWorkingDay bool
	StartTime    *string
	EndTime      *string
	Reason       *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewEmployeeScheduleOverride(
	employeeID string,
	date time.Time,
	isWorkingDay bool,
	startTime, endTime *string,
	reason *string,
) *EmployeeScheduleOverride {
	now := time.Now()
	return &EmployeeScheduleOverride{
		ID:           uuid.New().String(),
		EmployeeID:   employeeID,
		Date:         date,
		IsWorkingDay: isWorkingDay,
		StartTime:    startTime,
		EndTime:      endTime,
		Reason:       reason,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func ReconstituteEmployeeScheduleOverride(
	id, employeeID string,
	date time.Time,
	isWorkingDay bool,
	startTime, endTime *string,
	reason *string,
	createdAt, updatedAt time.Time,
) *EmployeeScheduleOverride {
	return &EmployeeScheduleOverride{
		ID:           id,
		EmployeeID:   employeeID,
		Date:         date,
		IsWorkingDay: isWorkingDay,
		StartTime:    startTime,
		EndTime:      endTime,
		Reason:       reason,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
