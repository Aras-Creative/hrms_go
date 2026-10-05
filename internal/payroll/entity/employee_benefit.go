package entity

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeBenefit struct {
	ID                string
	EmployeeID        string
	BenefitTypeID     string
	ParticipantNumber string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewEmployeeBenefit(
	employeeID string,
	benefitTypeID string,
	participantNumber string,
) *EmployeeBenefit {
	now := time.Now()
	return &EmployeeBenefit{
		ID:                uuid.New().String(),
		EmployeeID:        employeeID,
		BenefitTypeID:     benefitTypeID,
		ParticipantNumber: participantNumber,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func ReconstituteEmployeeBenefit(
	id string,
	employeeID string,
	benefitTypeID string,
	participantNumber string,
	createdAt time.Time,
	updatedAt time.Time,
) *EmployeeBenefit {
	return &EmployeeBenefit{
		ID:                id,
		EmployeeID:        employeeID,
		BenefitTypeID:     benefitTypeID,
		ParticipantNumber: participantNumber,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}
