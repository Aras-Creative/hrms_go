package entity

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type DeductionType struct {
	ID            string
	Name          string
	Slug          string
	Description   string
	DeductionType DeductionCalcType
	DefaultValue  float64
	ValueSource   ValueSource
	UnitAmount    Amount
	IsActive      bool
	IsMandatory   bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CalcContext carries the per-employee figures a payroll calculation depends on.
type CalcContext struct {
	BaseSalaryCents  int64
	WorkingDays      int
	UnpaidAbsentDays int
	AttendedDays     int
}

// DefaultWorkingDays mirrors the processor fallback when a work pattern yields no days.
const DefaultWorkingDays = 20

// DailyRate is the employee's salary spread over their working days. It is the
// value_source = daily_wage rate, and the fallback when no work pattern is defined.
func (c CalcContext) DailyRate() int64 {
	days := c.WorkingDays
	if days <= 0 {
		days = DefaultWorkingDays
	}
	return c.BaseSalaryCents / int64(days)
}

// EffectiveWorkingDays resolves the working-day divisor, guarding against division by zero.
func (c CalcContext) EffectiveWorkingDays() int {
	if c.WorkingDays <= 0 {
		return DefaultWorkingDays
	}
	return c.WorkingDays
}

func NewDeductionType(
	name string,
	slug string,
	description string,
	deductionType DeductionCalcType,
	defaultValue float64,
	valueSource ValueSource,
	unitAmount Amount,
	isMandatory bool,
) *DeductionType {
	now := time.Now()
	return &DeductionType{
		ID:            uuid.New().String(),
		Name:          name,
		Slug:          slug,
		Description:   description,
		DeductionType: deductionType,
		DefaultValue:  defaultValue,
		ValueSource:   valueSource,
		UnitAmount:    unitAmount,
		IsActive:      true,
		IsMandatory:   isMandatory,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func ReconstituteDeductionType(
	id string,
	name string,
	slug string,
	description string,
	deductionType string,
	defaultValue float64,
	valueSource string,
	unitAmountCents int64,
	isActive bool,
	isMandatory bool,
	createdAt time.Time,
	updatedAt time.Time,
) *DeductionType {
	return &DeductionType{
		ID:            id,
		Name:          name,
		Slug:          slug,
		Description:   description,
		DeductionType: DeductionCalcType(deductionType),
		DefaultValue:  defaultValue,
		ValueSource:   ValueSource(valueSource),
		UnitAmount:    AmountFromCents(unitAmountCents),
		IsActive:      isActive,
		IsMandatory:   isMandatory,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

// Calculate returns the deduction in cents. It is the authoritative formula shared by
// the payroll processor and the period overview, so a payslip always matches the
// summary shown before it is approved.
//
//	percentage : baseSalary * defaultValue / 100
//	fixed      : defaultValue, a flat amount for the whole period
//	per_day    : unpaidAbsentDays * dailyRate, where dailyRate comes from
//	             value_source = fixed      -> unitAmount
//	             value_source = daily_wage -> baseSalary / workingDays
func (dt *DeductionType) Calculate(c CalcContext) int64 {
	switch dt.DeductionType {
	case DeductionCalcPercentage:
		return int64(math.Round(float64(c.BaseSalaryCents) * ClampPercentage(dt.DefaultValue) / 100))
	case DeductionCalcPerDay:
		return int64(c.UnpaidAbsentDays) * dt.dailyRate(c)
	default:
		return int64(math.Round(dt.DefaultValue * 100))
	}
}

// PerDayRateCents exposes the resolved daily rate, used for payslip breakdowns.
func (dt *DeductionType) PerDayRateCents(c CalcContext) int64 {
	return dt.dailyRate(c)
}

func (dt *DeductionType) dailyRate(c CalcContext) int64 {
	if dt.ValueSource == ValueSourceDailyWage {
		return c.DailyRate()
	}
	return dt.UnitAmount.Cents()
}

// ValidateAssignmentValue checks a value being attached to an employee against the
// deduction type it will be read as.
//
// The setup endpoint cannot apply the rule with struct tags alone, because the accepted
// range depends on deduction_type, which lives on the master and is only known once the
// deduction_type_id is resolved. Without this, an amount in rupiah written into a
// percentage deduction is accepted silently and later deducted as a multiple of the
// salary: a 100000 "rupiah" entry became a 100000% deduction, which is 1000x the pay.
func (dt *DeductionType) ValidateAssignmentValue(value float64) error {
	if value < 0 {
		return fmt.Errorf("deduction value must be >= 0")
	}
	if dt.DeductionType != DeductionCalcPercentage {
		return nil
	}
	if value > 100 {
		return fmt.Errorf(
			"deduction %q is a percentage, so value is a percent of base salary and must be between 0 and 100; got %g. To charge a flat amount, use a deduction type of 'fixed'",
			dt.Name, value)
	}
	return nil
}
