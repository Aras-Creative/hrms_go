package models

import (
	"time"

	"hrms/internal/payroll/entity"
)

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
	CalcType           string
	EffectiveDate      time.Time
	EndDate            *time.Time
}

type UpdateEmployeeCompensationInput struct {
	CompensationItemID string
	Amount             float64
	Frequency          string
	CalcType           string
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
	ValueSource   string
	UnitAmount    float64
	IsMandatory   bool
}

type UpdateDeductionTypeInput struct {
	Name          *string
	Slug          *string
	Description   *string
	DeductionType *string
	DefaultValue  *float64
	ValueSource   *string
	UnitAmount    *float64
	IsActive      *bool
	IsMandatory   *bool
}

// --- Employee Deduction ---

type CreateEmployeeDeductionInput struct {
	EmployeeID      string
	DeductionTypeID string
	Value           *float64
	UnitAmount      *float64
	EffectiveDate   time.Time
	EndDate         *time.Time
}

type UpdateEmployeeDeductionInput struct {
	DeductionTypeID string
	Value           *float64
	UnitAmount      *float64
	EffectiveDate   time.Time
	EndDate         *time.Time
}

// --- Setup ---

type SetupCompensationItem struct {
	CompensationItemID string  `json:"compensation_item_id"`
	Amount             float64 `json:"amount"`
	Frequency          string  `json:"frequency"`
	CalcType           string  `json:"calc_type,omitempty"`
}

type SetupBenefitItem struct {
	BenefitTypeID     string `json:"benefit_type_id"`
	ParticipantNumber string `json:"participant_number"`
}

type SetupDeductionItem struct {
	DeductionTypeID string   `json:"deduction_type_id"`
	Value           *float64 `json:"value,omitempty"`
	UnitAmount      *float64 `json:"unit_amount,omitempty"`
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

// ManualPaySlipInput carries the figures for a hand-written slip.
type ManualPaySlipInput struct {
	PeriodID      string
	EmployeeID    string
	BaseSalary    float64
	Currency      string
	Compensations []ManualCompensationInput
	Deductions    []ManualDeductionInput
	AbsentDays    int
	// IncomeInputs are the descriptive, non-money figures typed onto the slip. They are
	// never summed into any total; see entity.IncomeInput for why.
	IncomeInputs []entity.IncomeInput
}

// UpdatePaySlipInput is a partial edit: a nil slice pointer means "leave unchanged",
// a non-nil (possibly empty) slice means "replace with exactly this".
type UpdatePaySlipInput struct {
	PaySlipID     string
	BaseSalary    *float64
	Compensations *[]ManualCompensationInput
	Deductions    *[]ManualDeductionInput
	AbsentDays    *int
	Recalculate   bool
	// IncomeInputs follows the same nil-means-unchanged rule as the breakdowns: nil leaves
	// the stored figures alone, a non-nil slice replaces the whole set so an omitted key
	// clears the figure rather than lingering on the slip.
	IncomeInputs *[]entity.IncomeInput
}

type ManualCompensationInput struct {
	CompensationItemID string
	Amount             float64
}

type ManualDeductionInput struct {
	DeductionTypeID string
	Amount          float64
}
