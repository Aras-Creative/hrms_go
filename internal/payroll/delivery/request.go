package delivery

import (
	"encoding/json"
	"strconv"

	"hrms/internal/payroll/entity"
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
	CompensationItemID string       `json:"compensation_item_id" validate:"required,uuid"`
	Amount             *FlexFloat64 `json:"amount,omitempty"`
	Frequency          string       `json:"frequency,omitempty" validate:"omitempty,oneof=monthly yearly one_time"`
	CalcType           string       `json:"calc_type,omitempty" validate:"omitempty,oneof=fixed per_attended_day"`
	EffectiveDate      string       `json:"effective_date" validate:"required"`
	EndDate            *string      `json:"end_date,omitempty"`
}

type SetupBenefitRequest struct {
	BenefitTypeID     string  `json:"benefit_type_id" validate:"required,uuid"`
	ParticipantNumber string  `json:"participant_number,omitempty"`
	EffectiveDate     string  `json:"effective_date" validate:"required"`
	EndDate           *string `json:"end_date,omitempty"`
}

type SetupDeductionRequest struct {
	DeductionTypeID string       `json:"deduction_type_id" validate:"required,uuid"`
	Value           *FlexFloat64 `json:"value,omitempty"`
	UnitAmount      *FlexFloat64 `json:"unit_amount,omitempty"`
	EffectiveDate   string       `json:"effective_date" validate:"required"`
	EndDate         *string      `json:"end_date,omitempty"`
}

type SetupBaseSalaryRequest struct {
	Amount        FlexFloat64 `json:"amount" validate:"required"`
	Currency      string      `json:"currency" validate:"omitempty,len=3"`
	EffectiveDate string      `json:"effective_date" validate:"required"`
	EndDate       *string     `json:"end_date,omitempty"`
	Notes         string      `json:"notes"`
}

type SetupEmployeePayrollRequest struct {
	EmployeeID    string                     `json:"employee_id" validate:"required,uuid"`
	BaseSalary    *SetupBaseSalaryRequest    `json:"base_salary,omitempty"`
	Compensations []SetupCompensationRequest `json:"compensations,omitempty"`
	Benefits      []SetupBenefitRequest      `json:"benefits,omitempty"`
	Deductions    []SetupDeductionRequest    `json:"deductions,omitempty"`
}

type CreateBaseSalaryRequest struct {
	EmployeeID    string      `json:"employee_id" validate:"required,uuid"`
	Amount        FlexFloat64 `json:"amount" validate:"required"`
	Currency      string      `json:"currency" validate:"omitempty,len=3"`
	EffectiveDate string      `json:"effective_date" validate:"required"`
	EndDate       *string     `json:"end_date" validate:"omitempty"`
	Notes         string      `json:"notes"`
}

type UpdateBaseSalaryRequest struct {
	Amount        FlexFloat64 `json:"amount" validate:"required"`
	Currency      string      `json:"currency" validate:"omitempty,len=3"`
	EffectiveDate string      `json:"effective_date" validate:"required"`
	EndDate       *string     `json:"end_date" validate:"omitempty"`
	Notes         string      `json:"notes"`
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
	Name                      string      `json:"name" validate:"required,min=1,max=255"`
	Description               string      `json:"description"`
	EmployerContributionType  string      `json:"employer_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployerContributionValue FlexFloat64 `json:"employer_contribution_value"`
	EmployeeContributionType  string      `json:"employee_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployeeContributionValue FlexFloat64 `json:"employee_contribution_value"`
}

type UpdateBenefitTypeRequest struct {
	Name                      string      `json:"name" validate:"required,min=1,max=255"`
	Description               string      `json:"description"`
	EmployerContributionType  string      `json:"employer_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployerContributionValue FlexFloat64 `json:"employer_contribution_value"`
	EmployeeContributionType  string      `json:"employee_contribution_type" validate:"required,oneof=percentage fixed"`
	EmployeeContributionValue FlexFloat64 `json:"employee_contribution_value"`
	IsActive                  *bool       `json:"is_active"`
}

type CreateDeductionTypeRequest struct {
	Name          string      `json:"name" validate:"required,min=1,max=255"`
	Description   string      `json:"description"`
	DeductionType string      `json:"deduction_type" validate:"required,oneof=percentage fixed per_day"`
	ValueSource   string      `json:"value_source" validate:"omitempty,oneof=fixed daily_wage"`
	UnitAmount    float64     `json:"unit_amount"`
	DefaultValue  FlexFloat64 `json:"default_value"`
	IsMandatory   bool        `json:"is_mandatory"`
}

type UpdateDeductionTypeRequest struct {
	Name          *string      `json:"name" validate:"omitempty,min=1,max=255"`
	Slug          *string      `json:"slug" validate:"omitempty,min=1,max=255"`
	Description   *string      `json:"description"`
	DeductionType *string      `json:"deduction_type" validate:"omitempty,oneof=percentage fixed per_day"`
	ValueSource   *string      `json:"value_source" validate:"omitempty,oneof=fixed daily_wage"`
	DefaultValue  *FlexFloat64 `json:"default_value"`
	UnitAmount    *FlexFloat64 `json:"unit_amount"`
	IsActive      *bool        `json:"is_active"`
	IsMandatory   *bool        `json:"is_mandatory"`
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

// IncomeInputRequest is one non-money figure HR records. Keys are limited to
// jumlah_sukses, persentase_rts and closing_bersih; anything else is rejected so a
// typo cannot create an input that silently never appears on the slip.
type IncomeInputRequest struct {
	Key   string      `json:"key" validate:"required"`
	Value FlexFloat64 `json:"value"`
	Unit  string      `json:"unit" validate:"omitempty,oneof=number percent currency days text"`
	Notes string      `json:"notes"`
}

type UpsertIncomeInputsRequest struct {
	Inputs []IncomeInputRequest `json:"inputs" validate:"required,dive"`
}

// toEntity leaves the key and unit as typed so entity validation can report the caller's
// own wording back to them instead of a normalised one.
func (r IncomeInputRequest) toEntity() entity.IncomeInput {
	return entity.IncomeInput{
		Key:   r.Key,
		Value: float64(r.Value),
		Unit:  entity.IncomeInputUnit(r.Unit),
		Notes: r.Notes,
	}
}

type CreateManualPaySlipRequest struct {
	EmployeeID    string                      `json:"employee_id" validate:"required,uuid"`
	BaseSalary    FlexFloat64                 `json:"base_salary" validate:"required"`
	Currency      string                      `json:"currency"`
	Compensations []ManualCompensationRequest `json:"compensations,omitempty"`
	Deductions    []ManualDeductionRequest    `json:"deductions,omitempty"`
	AbsentDays    int                         `json:"absent_days"`
	// IncomeInputs are the descriptive, non-money figures for this slip. Optional: a slip
	// without them is valid, and they never affect net salary.
	IncomeInputs []IncomeInputRequest `json:"income_inputs,omitempty"`
}

// UpdatePaySlipRequest uses pointer fields so an omitted key keeps the stored value
// while an explicit empty array clears the breakdown. Without this distinction a
// partial edit would silently zero out compensations or deductions.
type UpdatePaySlipRequest struct {
	BaseSalary    *FlexFloat64                 `json:"base_salary"`
	Compensations *[]ManualCompensationRequest `json:"compensations"`
	Deductions    *[]ManualDeductionRequest    `json:"deductions"`
	AbsentDays    *int                         `json:"absent_days"`
	Recalculate   bool                         `json:"recalculate"`
	// IncomeInputs holds the non-money figures for this slip, e.g. how many closings a
	// salesperson made. Sending the field replaces the whole set, so an omitted key clears
	// the figure. None of these values affect net salary.
	IncomeInputs *[]IncomeInputRequest `json:"income_inputs"`
}
