package entity

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeBaseSalary struct {
	ID            string
	EmployeeID    string
	Amount        Amount
	Currency      Currency
	EffectiveDate time.Time
	EndDate       *time.Time
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewEmployeeBaseSalary(
	employeeID string,
	amount Amount,
	currency Currency,
	effectiveDate time.Time,
	endDate *time.Time,
	notes string,
) *EmployeeBaseSalary {
	now := time.Now()
	return &EmployeeBaseSalary{
		ID:            uuid.New().String(),
		EmployeeID:    employeeID,
		Amount:        amount,
		Currency:      currency,
		EffectiveDate: effectiveDate,
		EndDate:       endDate,
		Notes:         notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func ReconstituteEmployeeBaseSalary(
	id string,
	employeeID string,
	amountCents int64,
	currency string,
	effectiveDate time.Time,
	endDate *time.Time,
	notes string,
	createdAt time.Time,
	updatedAt time.Time,
) *EmployeeBaseSalary {
	return &EmployeeBaseSalary{
		ID:            id,
		EmployeeID:    employeeID,
		Amount:        AmountFromCents(amountCents),
		Currency:      CurrencyFromDB(currency),
		EffectiveDate: effectiveDate,
		EndDate:       endDate,
		Notes:         notes,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}
