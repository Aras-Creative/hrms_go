package entity

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeCompensation struct {
	ID                 string
	EmployeeID         string
	CompensationItemID string
	Amount             Amount
	Frequency          Frequency
	EffectiveDate      time.Time
	EndDate            *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewEmployeeCompensation(
	employeeID string,
	compensationItemID string,
	amount Amount,
	frequency Frequency,
	effectiveDate time.Time,
	endDate *time.Time,
) *EmployeeCompensation {
	now := time.Now()
	return &EmployeeCompensation{
		ID:                 uuid.New().String(),
		EmployeeID:         employeeID,
		CompensationItemID: compensationItemID,
		Amount:             amount,
		Frequency:          frequency,
		EffectiveDate:      effectiveDate,
		EndDate:            endDate,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func ReconstituteEmployeeCompensation(
	id string,
	employeeID string,
	compensationItemID string,
	amountCents int64,
	frequency string,
	effectiveDate time.Time,
	endDate *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *EmployeeCompensation {
	return &EmployeeCompensation{
		ID:                 id,
		EmployeeID:         employeeID,
		CompensationItemID: compensationItemID,
		Amount:             AmountFromCents(amountCents),
		Frequency:          Frequency(frequency),
		EffectiveDate:      effectiveDate,
		EndDate:            endDate,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
}
