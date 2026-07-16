package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EmployeeWorkPattern struct {
	ID            string
	EmployeeID    string
	WorkPatternID string
	ValidFrom     time.Time
	ValidTo       *time.Time
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewEmployeeWorkPattern(employeeID, workPatternID string, validFrom time.Time, validTo *time.Time) *EmployeeWorkPattern {
	now := time.Now()
	return &EmployeeWorkPattern{
		ID:            uuid.New().String(),
		EmployeeID:    employeeID,
		WorkPatternID: workPatternID,
		ValidFrom:     validFrom,
		ValidTo:       validTo,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func ReconstituteEmployeeWorkPattern(
	id, employeeID, workPatternID string,
	validFrom time.Time,
	validTo *time.Time,
	isActive bool,
	createdAt, updatedAt time.Time,
) *EmployeeWorkPattern {
	return &EmployeeWorkPattern{
		ID:            id,
		EmployeeID:    employeeID,
		WorkPatternID: workPatternID,
		ValidFrom:     validFrom,
		ValidTo:       validTo,
		IsActive:      isActive,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func (e *EmployeeWorkPattern) Deactivate() error {
	if !e.IsActive {
		return fmt.Errorf("work pattern is already inactive")
	}
	e.IsActive = false
	e.UpdatedAt = time.Now()
	return nil
}

func (e *EmployeeWorkPattern) SetValidTo(t time.Time) error {
	if t.Before(e.ValidFrom) {
		return fmt.Errorf("valid_to must not be before valid_from")
	}
	e.ValidTo = &t
	e.UpdatedAt = time.Now()
	return nil
}
