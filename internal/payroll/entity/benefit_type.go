package entity

import (
	"time"

	"github.com/google/uuid"
)

type BenefitType struct {
	ID                        string
	Name                      string
	Description               string
	EmployerContributionType  ContributionType
	EmployerContributionValue float64
	EmployeeContributionType  ContributionType
	EmployeeContributionValue float64
	IsActive                  bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

func NewBenefitType(
	name string,
	description string,
	employerContributionType ContributionType,
	employerContributionValue float64,
	employeeContributionType ContributionType,
	employeeContributionValue float64,
) *BenefitType {
	now := time.Now()
	return &BenefitType{
		ID:                        uuid.New().String(),
		Name:                      name,
		Description:               description,
		EmployerContributionType:  employerContributionType,
		EmployerContributionValue: employerContributionValue,
		EmployeeContributionType:  employeeContributionType,
		EmployeeContributionValue: employeeContributionValue,
		IsActive:                  true,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
}

func ReconstituteBenefitType(
	id string,
	name string,
	description string,
	employerContributionType string,
	employerContributionValue float64,
	employeeContributionType string,
	employeeContributionValue float64,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) *BenefitType {
	return &BenefitType{
		ID:                        id,
		Name:                      name,
		Description:               description,
		EmployerContributionType:  ContributionType(employerContributionType),
		EmployerContributionValue: employerContributionValue,
		EmployeeContributionType:  ContributionType(employeeContributionType),
		EmployeeContributionValue: employeeContributionValue,
		IsActive:                  isActive,
		CreatedAt:                 createdAt,
		UpdatedAt:                 updatedAt,
	}
}
