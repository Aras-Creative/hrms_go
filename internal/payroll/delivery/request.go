package delivery

import (
	"encoding/json"
	"strconv"
)

// FlexFloat64 accepts both JSON number and JSON string for float64 fields.
type FlexFloat64 float64

func (f *FlexFloat64) UnmarshalJSON(data []byte) error {
	var v float64
	if err := json.Unmarshal(data, &v); err == nil {
		*f = FlexFloat64(v)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*f = FlexFloat64(v)
	return nil
}

type SetupCompensationRequest struct {
	CompensationItemID string      `json:"compensation_item_id" validate:"required,uuid"`
	Amount             *FlexFloat64 `json:"amount,omitempty"`
	Frequency          string      `json:"frequency,omitempty"`
	EffectiveDate      string      `json:"effective_date" validate:"required"`
	EndDate            *string     `json:"end_date,omitempty"`
}

type SetupBenefitRequest struct {
	BenefitTypeID     string  `json:"benefit_type_id" validate:"required,uuid"`
	ParticipantNumber string  `json:"participant_number,omitempty"`
	EffectiveDate     string  `json:"effective_date" validate:"required"`
	EndDate           *string `json:"end_date,omitempty"`
}

type SetupDeductionRequest struct {
	DeductionTypeID string      `json:"deduction_type_id" validate:"required,uuid"`
	Value           *FlexFloat64 `json:"value,omitempty"`
	EffectiveDate   string      `json:"effective_date" validate:"required"`
	EndDate         *string     `json:"end_date,omitempty"`
}

type SetupBaseSalaryRequest struct {
	Amount        FlexFloat64 `json:"amount" validate:"required"`
	Currency      string      `json:"currency" validate:"omitempty,len=3"`
	EffectiveDate string      `json:"effective_date" validate:"required"`
	EndDate       *string     `json:"end_date,omitempty"`
	Notes         string      `json:"notes"`
}

type SetupEmployeePayrollRequest struct {
	EmployeeID    string                    `json:"employee_id" validate:"required,uuid"`
	BaseSalary    *SetupBaseSalaryRequest   `json:"base_salary,omitempty"`
	Compensations []SetupCompensationRequest `json:"compensations,omitempty"`
	Benefits      []SetupBenefitRequest     `json:"benefits,omitempty"`
	Deductions    []SetupDeductionRequest   `json:"deductions,omitempty"`
}

type CreateBaseSalaryRequest struct {
	EmployeeID    string     `json:"employee_id" validate:"required,uuid"`
	Amount        FlexFloat64 `json:"amount" validate:"required"`
	Currency      string     `json:"currency" validate:"omitempty,len=3"`
	EffectiveDate string     `json:"effective_date" validate:"required"`
	EndDate       *string    `json:"end_date" validate:"omitempty"`
	Notes         string     `json:"notes"`
}

type UpdateBaseSalaryRequest struct {
	Amount        FlexFloat64 `json:"amount" validate:"required"`
	Currency      string     `json:"currency" validate:"omitempty,len=3"`
	EffectiveDate string     `json:"effective_date" validate:"required"`
	EndDate       *string    `json:"end_date" validate:"omitempty"`
	Notes         string     `json:"notes"`
}

type CreateCompensationItemRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	ItemType    string `json:"item_type" validate:"required,oneof=recurring one_time"`
	Description string `json:"description"`
	IsTaxable   bool   `json:"is_taxable"`
}

type UpdateCompensationItemRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	ItemType    string `json:"item_type" validate:"required,oneof=recurring one_time"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
	IsTaxable   *bool  `json:"is_taxable"`
}

type CreateBenefitTypeRequest struct {
	Name                      string     `json:"name" validate:"required,min=1,max=255"`
	Description               string     `json:"description"`
	EmployerContributionType  string     `json:"employer_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployerContributionValue FlexFloat64 `json:"employer_contribution_value"`
	EmployeeContributionType  string     `json:"employee_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployeeContributionValue FlexFloat64 `json:"employee_contribution_value"`
}

type UpdateBenefitTypeRequest struct {
	Name                      string     `json:"name" validate:"required,min=1,max=255"`
	Description               string     `json:"description"`
	EmployerContributionType  string     `json:"employer_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployerContributionValue FlexFloat64 `json:"employer_contribution_value"`
	EmployeeContributionType  string     `json:"employee_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployeeContributionValue FlexFloat64 `json:"employee_contribution_value"`
	IsActive                  *bool      `json:"is_active"`
}

type CreateDeductionTypeRequest struct {
	Name          string      `json:"name" validate:"required,min=1,max=255"`
	Description   string      `json:"description"`
	DeductionType string      `json:"deduction_type" validate:"required,oneof=percentage fixed"`
	DefaultValue  FlexFloat64 `json:"default_value"`
	IsMandatory   bool        `json:"is_mandatory"`
}

type UpdateDeductionTypeRequest struct {
	Name          string      `json:"name" validate:"required,min=1,max=255"`
	Description   string      `json:"description"`
	DeductionType string      `json:"deduction_type" validate:"required,oneof=percentage fixed"`
	DefaultValue  FlexFloat64 `json:"default_value"`
	IsActive      *bool       `json:"is_active"`
	IsMandatory   *bool       `json:"is_mandatory"`
}

// --- Period ---

type CreatePeriodRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=255"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}

type UpdatePeriodRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=255"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}

// --- Manual Pay Slip ---

type ManualCompensationRequest struct {
	CompensationItemID string      `json:"compensation_item_id" validate:"required,uuid"`
	Amount             FlexFloat64 `json:"amount" validate:"required"`
}

type ManualDeductionRequest struct {
	DeductionTypeID string      `json:"deduction_type_id" validate:"required,uuid"`
	Amount          FlexFloat64 `json:"amount" validate:"required"`
}

type CreateManualPaySlipRequest struct {
	EmployeeID      string                     `json:"employee_id" validate:"required,uuid"`
	BaseSalary      FlexFloat64                `json:"base_salary" validate:"required"`
	Currency        string                     `json:"currency"`
	Compensations   []ManualCompensationRequest `json:"compensations,omitempty"`
	Deductions      []ManualDeductionRequest    `json:"deductions,omitempty"`
	AbsentDeduction FlexFloat64                `json:"absent_deduction"`
	AbsentDays      int                        `json:"absent_days"`
}
