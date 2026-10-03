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
	// UnitAmount overrides the per-day rate carried by the deduction type. Only
	// meaningful when the type is per_day; nil falls back to the type's unit_amount.
	UnitAmount *int64
	// Basis fields below are denormalised from deduction_types on read. An assignment
	// only stores deduction_type_id, so without them a client cannot tell whether value
	// or unit_amount is the field that matters, and has to fetch every type separately.
	DeductionType     DeductionCalcType
	DeductionTypeName string
	ValueSource       ValueSource
	EffectiveDate     time.Time
	EndDate           *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewEmployeeDeduction(
	employeeID string,
	deductionTypeID string,
	value *float64,
	unitAmount *int64,
	effectiveDate time.Time,
	endDate *time.Time,
) *EmployeeDeduction {
	now := time.Now()
	return &EmployeeDeduction{
		ID:              uuid.New().String(),
		EmployeeID:      employeeID,
		DeductionTypeID: deductionTypeID,
		Value:           value,
		UnitAmount:      unitAmount,
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
	unitAmount *int64,
	effectiveDate time.Time,
	endDate *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deductionType DeductionCalcType,
	deductionTypeName string,
	valueSource ValueSource,
) *EmployeeDeduction {
	return &EmployeeDeduction{
		ID:                id,
		EmployeeID:        employeeID,
		DeductionTypeID:   deductionTypeID,
		Value:             value,
		UnitAmount:        unitAmount,
		EffectiveDate:     effectiveDate,
		EndDate:           endDate,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		DeductionType:     deductionType,
		DeductionTypeName: deductionTypeName,
		ValueSource:       valueSource,
	}
}
