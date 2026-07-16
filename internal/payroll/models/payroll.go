package models

import "time"

// --- Base Salary ---

type CreateBaseSalaryInput struct {
	EmployeeID    string
	Amount        float64
	Currency      string
	EffectiveDate time.Time
	EndDate       *time.Time
	Notes         string
}

type UpdateBaseSalaryInput struct {
	Amount        float64
	Currency      string
	EffectiveDate time.Time
	EndDate       *time.Time
	Notes         string
}

// --- Compensation Item ---

type CreateCompensationItemInput struct {
	Name        string
	ItemType    string
	Description string
	IsTaxable   bool
}

type UpdateCompensationItemInput struct {
	Name        string
	ItemType    string
	Description string
	IsActive    *bool
	IsTaxable   *bool
}

// --- Employee Compensation ---

type CreateEmployeeCompensationInput struct {
	EmployeeID         string
	CompensationItemID string
	Amount             float64
	Frequency          string
	EffectiveDate      time.Time
	EndDate            *time.Time
}

type UpdateEmployeeCompensationInput struct {
	CompensationItemID string
	Amount             float64
	Frequency          string
	EffectiveDate      time.Time
	EndDate            *time.Time
}

// --- Benefit Type ---

type CreateBenefitTypeInput struct {
	Name                      string
	Description               string
	EmployerContributionType  string
	EmployerContributionValue float64
	EmployeeContributionType  string
	EmployeeContributionValue float64
}

type UpdateBenefitTypeInput struct {
	Name                      string
	Description               string
	EmployerContributionType  string
	EmployerContributionValue float64
	EmployeeContributionType  string
	EmployeeContributionValue float64
	IsActive                  *bool
}

// --- Employee Benefit ---

type CreateEmployeeBenefitInput struct {
	EmployeeID        string
	BenefitTypeID     string
	ParticipantNumber string
	EffectiveDate     time.Time
	EndDate           *time.Time
}

type UpdateEmployeeBenefitInput struct {
	BenefitTypeID     string
	ParticipantNumber string
	EffectiveDate     time.Time
	EndDate           *time.Time
}

// --- Deduction Type ---

type CreateDeductionTypeInput struct {
	Name          string
	Slug          string
	Description   string
	DeductionType string
	DefaultValue  float64
	IsMandatory   bool
}

type UpdateDeductionTypeInput struct {
	Name          string
	Description   string
	DeductionType string
	DefaultValue  float64
	IsActive      *bool
	IsMandatory   *bool
}

// --- Employee Deduction ---

type CreateEmployeeDeductionInput struct {
	EmployeeID      string
	DeductionTypeID string
	Value           *float64
	EffectiveDate   time.Time
	EndDate         *time.Time
}

type UpdateEmployeeDeductionInput struct {
	DeductionTypeID string
	Value           *float64
	EffectiveDate   time.Time
	EndDate         *time.Time
}

// --- Setup ---

type SetupCompensationItem struct {
	CompensationItemID string    `json:"compensation_item_id"`
	Amount             float64   `json:"amount"`
	Frequency          string    `json:"frequency"`
	EffectiveDate      time.Time `json:"effective_date"`
	EndDate            *time.Time `json:"end_date,omitempty"`
}

type SetupBenefitItem struct {
	BenefitTypeID     string    `json:"benefit_type_id"`
	ParticipantNumber string    `json:"participant_number"`
	EffectiveDate     time.Time `json:"effective_date"`
	EndDate           *time.Time `json:"end_date,omitempty"`
}

type SetupDeductionItem struct {
	DeductionTypeID string    `json:"deduction_type_id"`
	Value           *float64   `json:"value,omitempty"`
	EffectiveDate   time.Time `json:"effective_date"`
	EndDate         *time.Time `json:"end_date,omitempty"`
}

type SetupEmployeePayrollInput struct {
	EmployeeID    string
	BaseSalary    *SetupBaseSalary
	Compensations []SetupCompensationItem
	Benefits      []SetupBenefitItem
	Deductions    []SetupDeductionItem
}

type SetupBaseSalary struct {
	Amount        float64
	Currency      string
	EffectiveDate time.Time
	EndDate       *time.Time
	Notes         string
}

// --- Period ---

type CreatePeriodInput struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

type UpdatePeriodInput struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

type ManualPaySlipInput struct {
	PeriodID        string
	EmployeeID      string
	BaseSalary      float64
	Currency        string
	Compensations   []ManualCompensationInput
	Deductions      []ManualDeductionInput
	AbsentDeduction float64
	AbsentDays      int
}

type ManualCompensationInput struct {
	CompensationItemID string
	Amount             float64
}

type ManualDeductionInput struct {
	DeductionTypeID string
	Amount          float64
}
