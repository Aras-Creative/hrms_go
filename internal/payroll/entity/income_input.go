package entity

import (
	"fmt"
	"strings"
)

// IncomeInputUnit is a display hint for an income input. It never affects a calculation:
// nothing in the payslip builder or the calculation package reads Unit. It exists so the
// API and the PDF can print "5 %" instead of a bare 5.
type IncomeInputUnit string

const (
	IncomeInputUnitNumber   IncomeInputUnit = "number"
	IncomeInputUnitPercent  IncomeInputUnit = "percent"
	IncomeInputUnitCurrency IncomeInputUnit = "currency"
	IncomeInputUnitDays     IncomeInputUnit = "days"
	IncomeInputUnitText     IncomeInputUnit = "text"
)

// incomeInputLabels gives well-known keys a friendlier display name. The key stays the
// stored identifier and the label is derived, never stored, so renaming a figure later is a
// code change and not a data migration. Both the PDF and the API use it, so the slip and the
// client cannot drift into calling the same thing different names. Keys are free-form, so
// this table is only a convenience for the keys the CRM payslip is known to use; any other
// key simply falls back to itself.
var incomeInputLabels = map[string]string{
	"jumlah_sukses":  "Jumlah Sukses",
	"persentase_rts": "Persentase RTS",
	"closing_bersih": "Closing Bersih",
}

// IncomeInputLabel returns the display name for a key. A key without a known label falls
// back to the key itself rather than to an empty string: a blank label would render an
// unlabelled line on the slip, while the key at least tells a reader what was entered.
func IncomeInputLabel(key string) string {
	if l, ok := incomeInputLabels[key]; ok {
		return l
	}
	return key
}

// IncomeInput is one non-money figure HR records on a pay slip, for example how many
// successful closings a salesperson made or the rate that applied.
//
// These deliberately never reach the payslip totals. Every line in
// compensations_breakdown is summed into net salary, so a count like "10 closings" would
// add Rp 10 to the employee's pay. Payroll setup stays financial; these are the
// descriptive half, stored apart from the breakdown so they can never be added by accident.
type IncomeInput struct {
	Key   string          `json:"key"`
	Value float64         `json:"value"`
	Unit  IncomeInputUnit `json:"unit"`
	Notes string          `json:"notes,omitempty"`
}

func ParseIncomeInputUnit(s string) (IncomeInputUnit, error) {
	if s == "" {
		return IncomeInputUnitNumber, nil
	}
	u := IncomeInputUnit(strings.ToLower(strings.TrimSpace(s)))
	switch u {
	case IncomeInputUnitNumber, IncomeInputUnitPercent, IncomeInputUnitCurrency,
		IncomeInputUnitDays, IncomeInputUnitText:
		return u, nil
	}
	return "", fmt.Errorf("invalid income input unit %q: expected number, percent, currency, days or text", s)
}

// ParseIncomeInputKey normalises a free-form key. Any non-empty key is accepted so HR can
// record a figure the CRM payroll did not anticipate; only surrounding whitespace is trimmed
// and a length cap keeps a stray paste from filling the column.
func ParseIncomeInputKey(s string) (string, error) {
	k := strings.TrimSpace(s)
	if k == "" {
		return "", fmt.Errorf("income input key is required")
	}
	if len(k) > 100 {
		return "", fmt.Errorf("income input key must be 100 characters or fewer")
	}
	return k, nil
}

// NewIncomeInput validates one submitted figure and normalises it. Everything that can be
// wrong with a hand-entered input is caught here, before the caller writes anything.
func NewIncomeInput(key string, value float64, unit IncomeInputUnit, notes string) (IncomeInput, error) {
	k, err := ParseIncomeInputKey(key)
	if err != nil {
		return IncomeInput{}, err
	}
	u, err := ParseIncomeInputUnit(string(unit))
	if err != nil {
		return IncomeInput{}, err
	}
	return IncomeInput{
		Key:   k,
		Value: value,
		Unit:  u,
		Notes: strings.TrimSpace(notes),
	}, nil
}

// ValidateIncomeInputs normalises a whole submitted set and rejects a repeated key. The
// input order is preserved so the payslip's calculation notes render in the order the
// figures were entered, rather than being reshuffled into key order.
//
// The returned slice is non-nil even when empty: the caller stores "[]" rather than null
// so the column never carries a null document that readers then have to special-case.
func ValidateIncomeInputs(raw []IncomeInput) ([]IncomeInput, error) {
	seen := make(map[string]bool, len(raw))
	out := make([]IncomeInput, 0, len(raw))
	for _, in := range raw {
		normalised, err := NewIncomeInput(in.Key, in.Value, in.Unit, in.Notes)
		if err != nil {
			return nil, err
		}
		if seen[normalised.Key] {
			return nil, fmt.Errorf("duplicate income input key %q", normalised.Key)
		}
		seen[normalised.Key] = true
		out = append(out, normalised)
	}
	return out, nil
}
