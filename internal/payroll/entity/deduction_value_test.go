package entity

import (
	"strings"
	"testing"
)

// The reported failure was a 100000 "rupiah" amount written into a percentage deduction,
// which Calculate then applied as 100000% of the salary. These tests pin both halves of
// the fix: the value is rejected on the way in, and clamped on the way out for rows that
// already exist.
func TestValidateAssignmentValueRejectsOutOfRangePercentage(t *testing.T) {
	dt := &DeductionType{Name: "Potongan BPJS Kesehatan", DeductionType: DeductionCalcPercentage}

	err := dt.ValidateAssignmentValue(100000)
	if err == nil {
		t.Fatal("100000 in a percentage deduction must be rejected, not stored")
	}
	// The message has to explain the unit, or the caller just retries with the same number.
	for _, want := range []string{"percentage", "0 and 100", "fixed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q so the caller knows the unit", err.Error(), want)
		}
	}
}

func TestValidateAssignmentValueAcceptsValidPercentage(t *testing.T) {
	dt := &DeductionType{Name: "BPJS", DeductionType: DeductionCalcPercentage}
	for _, v := range []float64{0, 1, 4.75, 100} {
		if err := dt.ValidateAssignmentValue(v); err != nil {
			t.Errorf("percentage %v should be accepted: %v", v, err)
		}
	}
}

// A flat deduction may legitimately exceed 100; only the percentage type is bounded.
func TestValidateAssignmentValueAllowsLargeFixedAmount(t *testing.T) {
	dt := &DeductionType{Name: "Kasbon", DeductionType: DeductionCalcFixed}
	if err := dt.ValidateAssignmentValue(5_000_000); err != nil {
		t.Errorf("a fixed deduction of 5,000,000 is legitimate: %v", err)
	}
	perDay := &DeductionType{Name: "Alpa", DeductionType: DeductionCalcPerDay}
	if err := perDay.ValidateAssignmentValue(106818.18); err != nil {
		t.Errorf("a per_day rate should be accepted: %v", err)
	}
}

func TestValidateAssignmentValueRejectsNegative(t *testing.T) {
	dt := &DeductionType{Name: "BPJS", DeductionType: DeductionCalcFixed}
	if err := dt.ValidateAssignmentValue(-1); err == nil {
		t.Error("a negative deduction value must be rejected")
	}
}

// The legacy row that produced the negative payslip: 100000% of Rp 2.350.000. Clamping
// bounds it to the whole salary instead of 1000x it.
func TestCalculateClampsLegacyPercentage(t *testing.T) {
	base := int64(235_000_000) // Rp 2.350.000 in cents

	legacy := &DeductionType{
		Name:          "Potongan BPJS Kesehatan",
		DeductionType: DeductionCalcPercentage,
		DefaultValue:  100000,
	}
	got := legacy.Calculate(CalcContext{BaseSalaryCents: base})
	if got != base {
		t.Errorf("Calculate() = %d, want the salary itself (%d) rather than a multiple of it", got, base)
	}
	if got < 0 || got > base {
		t.Errorf("a single deduction must never exceed the salary: got %d for base %d", got, base)
	}

	sane := &DeductionType{DeductionType: DeductionCalcPercentage, DefaultValue: 4}
	if got := sane.Calculate(CalcContext{BaseSalaryCents: base}); got != 9_400_000 {
		t.Errorf("4%% of %d = %d, want 9400000: clamping must not alter valid input", base, got)
	}
}

// Exactly 100% is a legitimate deduction, so the clamp must include the boundary.
func TestClampPercentageBoundaries(t *testing.T) {
	cases := map[float64]float64{-1: 0, 0: 0, 50: 50, 100: 100, 100.5: 100, 1e9: 100}
	for in, want := range cases {
		if got := ClampPercentage(in); got != want {
			t.Errorf("ClampPercentage(%v) = %v, want %v", in, got, want)
		}
	}
}
