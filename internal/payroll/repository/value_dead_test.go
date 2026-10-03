package repository

import (
	"testing"

	"hrms/internal/payroll/entity"
)

// TestValueIsDeadForPerDay documents that default_value / value play no part in the
// per_day branch: changing it must not move a cent.
func TestValueIsDeadForPerDay(t *testing.T) {
	ctx := entity.CalcContext{BaseSalaryCents: 1_000_000_000, WorkingDays: 22, UnpaidAbsentDays: 3}

	low := CalcDedRow{Type: "per_day", ValueSource: "daily_wage", Value: 0}
	high := CalcDedRow{Type: "per_day", ValueSource: "daily_wage", Value: 99}

	if low.CalculateCents(ctx) != high.CalculateCents(ctx) {
		t.Errorf("value moved a per_day deduction: %d vs %d", low.CalculateCents(ctx), high.CalculateCents(ctx))
	}

	// Same for the fixed value_source branch.
	fixedLow := CalcDedRow{Type: "per_day", ValueSource: "fixed", UnitAmount: 1_500_000, Value: 0}
	fixedHigh := CalcDedRow{Type: "per_day", ValueSource: "fixed", UnitAmount: 1_500_000, Value: 99}
	if fixedLow.CalculateCents(ctx) != fixedHigh.CalculateCents(ctx) {
		t.Error("value moved a per_day deduction with a fixed source")
	}

	// Sanity: value still drives percentage and fixed.
	pctLow := CalcDedRow{Type: "percentage", Value: 0}
	pctHigh := CalcDedRow{Type: "percentage", Value: 5}
	if pctLow.CalculateCents(ctx) == pctHigh.CalculateCents(ctx) {
		t.Error("value should still drive a percentage deduction")
	}
}
