package entity

import "testing"

// TestParseFrequencyRejectsCalcType guards the split introduced by migration 000048:
// a caller sending the old overloaded value must get an error, not a silent reinterpretation.
func TestParseFrequencyRejectsCalcType(t *testing.T) {
	for _, s := range []string{"per_attended_day", "per_day", "fixed"} {
		if _, err := ParseFrequency(s); err == nil {
			t.Errorf("ParseFrequency(%q) should fail: that column only carries cadence", s)
		}
	}

	for _, s := range []string{"monthly", "yearly", "one_time", "MONTHLY", "  yearly  "} {
		if _, err := ParseFrequency(s); err != nil {
			t.Errorf("ParseFrequency(%q) error = %v", s, err)
		}
	}
}

func TestParseCompensationCalcType(t *testing.T) {
	for _, s := range []string{"fixed", "per_attended_day", " PER_ATTENDED_DAY "} {
		got, err := ParseCompensationCalcType(s)
		if err != nil {
			t.Errorf("ParseCompensationCalcType(%q) error = %v", s, err)
		}
		if !got.IsValid() {
			t.Errorf("ParseCompensationCalcType(%q) = %q, not valid", s, got)
		}
	}

	// Cadence values must not be accepted as a derivation.
	for _, s := range []string{"monthly", "yearly", "one_time", ""} {
		if _, err := ParseCompensationCalcType(s); err == nil {
			t.Errorf("ParseCompensationCalcType(%q) should fail", s)
		}
	}
}
