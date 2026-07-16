package entity

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeDeduction struct {
	ID              string
	EmployeeID      string
	DeductionTypeID string
	Value           *float64
	EffectiveDate   time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewEmployeeDeduction(
	employeeID string,
	deductionTypeID string,
	value *float64,
	effectiveDate time.Time,
	endDate *time.Time,
) *EmployeeDeduction {
	now := time.Now()
	return &EmployeeDeduction{
		ID:              uuid.New().String(),
		EmployeeID:      employeeID,
		DeductionTypeID: deductionTypeID,
		Value:           value,
		EffectiveDate:   effectiveDate,
		EndDate:         endDate,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func ReconstituteEmployeeDeduction(
	id string,
	employeeID string,
	deductionTypeID string,
	value *float64,
	effectiveDate time.Time,
	endDate *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *EmployeeDeduction {
	return &EmployeeDeduction{
		ID:              id,
		EmployeeID:      employeeID,
		DeductionTypeID: deductionTypeID,
		Value:           value,
		EffectiveDate:   effectiveDate,
		EndDate:         endDate,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}
