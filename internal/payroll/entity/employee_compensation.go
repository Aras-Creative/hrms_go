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
	CalcType           CompensationCalcType
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewEmployeeCompensation(
	employeeID string,
	compensationItemID string,
	amount Amount,
	frequency Frequency,
	calcType CompensationCalcType,
) *EmployeeCompensation {
	now := time.Now()
	return &EmployeeCompensation{
		ID:                 uuid.New().String(),
		EmployeeID:         employeeID,
		CompensationItemID: compensationItemID,
		Amount:             amount,
		Frequency:          frequency,
		CalcType:           calcType,
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
	calcType string,
	createdAt time.Time,
	updatedAt time.Time,
) *EmployeeCompensation {
	return &EmployeeCompensation{
		ID:                 id,
		EmployeeID:         employeeID,
		CompensationItemID: compensationItemID,
		Amount:             AmountFromCents(amountCents),
		Frequency:          Frequency(frequency),
		CalcType:           CompensationCalcType(calcType),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
}
