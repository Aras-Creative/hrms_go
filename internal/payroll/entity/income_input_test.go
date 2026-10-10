package entity

import (
	"strings"
	"testing"
)

func TestParseIncomeInputKey(t *testing.T) {
	// Keys are free-form: anything non-empty is accepted and only trimmed.
	valid := []string{"jumlah_sukses", "persentase_rts", "closing_bersih", "persen_rts", "JUMLAH_SUKSES", "closing"}
	for _, k := range valid {
		got, err := ParseIncomeInputKey(k)
		if err != nil {
			t.Errorf("ParseIncomeInputKey(%q) unexpected error: %v", k, err)
		}
		if got != k {
			t.Errorf("ParseIncomeInputKey(%q) = %q", k, got)
		}
	}

	// Surrounding whitespace is normalised rather than rejected, so a pasted key still
	// lands on the same key.
	if got, err := ParseIncomeInputKey("  jumlah_sukses  "); err != nil || got != "jumlah_sukses" {
		t.Errorf("padded key should normalise, got %q err %v", got, err)
	}

	// Only a blank key is rejected, along with anything past the length cap.
	invalid := []string{"", "   ", strings.Repeat("a", 101)}
	for _, k := range invalid {
		if _, err := ParseIncomeInputKey(k); err == nil {
			t.Errorf("ParseIncomeInputKey(%q) should have failed", k)
		}
	}
}

func TestParseIncomeInputUnit(t *testing.T) {
	if u, err := ParseIncomeInputUnit(""); err != nil || u != IncomeInputUnitNumber {
		t.Errorf("empty unit should default to number, got %q err %v", u, err)
	}
	if u, err := ParseIncomeInputUnit(" PERCENT "); err != nil || u != IncomeInputUnitPercent {
		t.Errorf("PERCENT should normalise, got %q err %v", u, err)
	}
	if _, err := ParseIncomeInputUnit("kg"); err == nil {
		t.Error("kg should be rejected")
	}
}

func TestValidateIncomeInputsPreservesOrderAndNormalises(t *testing.T) {
	got, err := ValidateIncomeInputs([]IncomeInput{
		{Key: "closing_bersih", Value: 8, Unit: IncomeInputUnit(" NUMBER ")},
		{Key: " jumlah_sukses ", Value: 10, Unit: IncomeInputUnitPercent},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Input order is preserved so the calculation notes render as entered; only the key and
	// unit are normalised.
	if len(got) != 2 || got[0].Key != "closing_bersih" || got[1].Key != "jumlah_sukses" {
		t.Fatalf("input order not preserved: %+v", got)
	}
	if got[0].Unit != IncomeInputUnitNumber {
		t.Errorf("unit not normalised: %q", got[0].Unit)
	}
}

func TestValidateIncomeInputsRejectsDuplicates(t *testing.T) {
	_, err := ValidateIncomeInputs([]IncomeInput{
		{Key: "closing_bersih", Value: 8},
		{Key: "closing_bersih", Value: 9},
	})
	if err == nil {
		t.Fatal("a repeated key must be rejected, otherwise the second value silently wins")
	}
}

func TestValidateIncomeInputsEmptyIsNotNil(t *testing.T) {
	got, err := ValidateIncomeInputs(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("want a non-nil empty slice so the column stores [] rather than null")
	}
}
