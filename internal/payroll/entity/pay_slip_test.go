package entity

import "testing"

func TestPaySlipMarkManualRecalculates(t *testing.T) {
	ps := ReconstitutePaySlip(
		"ps-1", "period-1", "emp-1",
		5_000_000_00, 500_000_00, 100_000_00,
		0,
		5_400_000_00,
		"IDR", string(PaySlipSourceAuto),
		[]byte(`[{"compensation_item_id":"c1","name":"Tunjangan","amount":500000}]`),
		[]byte(`[{"deduction_type_id":"d1","name":"Potongan","amount":100000}]`),
		testTime(), testTime(),
	)

	if ps.Source != PaySlipSourceAuto {
		t.Fatalf("expected an auto slip, got %q", ps.Source)
	}

	ps.MarkManual()

	if ps.Source != PaySlipSourceManual {
		t.Errorf("Source = %q, want manual", ps.Source)
	}
	if ps.TotalCompensations.Cents() != 500_000_00 {
		t.Errorf("TotalCompensations = %d, want %d", ps.TotalCompensations.Cents(), 500_000_00)
	}
	if ps.TotalDeductions.Cents() != 100_000_00 {
		t.Errorf("TotalDeductions = %d, want %d", ps.TotalDeductions.Cents(), 100_000_00)
	}
	want := int64(5_000_000_00 + 500_000_00 - 100_000_00)
	if ps.NetSalary.Cents() != want {
		t.Errorf("NetSalary = %d, want %d", ps.NetSalary.Cents(), want)
	}
}

func TestPaySlipRecalculateTotalsAfterBreakdownEdit(t *testing.T) {
	ps := ReconstitutePaySlip(
		"ps-1", "period-1", "emp-1",
		5_000_000_00, 0, 0,
		0,
		5_000_000_00,
		"IDR", string(PaySlipSourceManual),
		nil, nil,
		testTime(), testTime(),
	)

	// An empty breakdown must clear the totals rather than leave stale values.
	ps.RecalculateTotals()
	if !ps.TotalCompensations.IsZero() || !ps.TotalDeductions.IsZero() {
		t.Errorf("expected zeroed totals, got comp=%v ded=%v", ps.TotalCompensations.Float(), ps.TotalDeductions.Float())
	}
	if ps.NetSalary.Cents() != 5_000_000_00 {
		t.Errorf("NetSalary = %d, want %d", ps.NetSalary.Cents(), 5_000_000_00)
	}

	ps.CompensationsBreakdown = []CompensationBreakdown{{Name: "Tunjangan Makan", Amount: 450_000}}
	ps.DeductionsBreakdown = []DeductionBreakdown{{Name: "Potongan Alpa", Amount: 150_000}}
	ps.RecalculateTotals()

	if ps.TotalCompensations.Cents() != 45_000_000 {
		t.Errorf("TotalCompensations = %d, want %d", ps.TotalCompensations.Cents(), 45_000_000)
	}
	if ps.TotalDeductions.Cents() != 15_000_000 {
		t.Errorf("TotalDeductions = %d, want %d", ps.TotalDeductions.Cents(), 15_000_000)
	}
	if want := int64(5_000_000_00 + 45_000_000 - 15_000_000); ps.NetSalary.Cents() != want {
		t.Errorf("NetSalary = %d, want %d", ps.NetSalary.Cents(), want)
	}
}

func TestPaySlipBuilderRecalculatesNetFromBreakdown(t *testing.T) {
	ps := NewPaySlipBuilder("period-1", "emp-1").
		WithBaseSalary(AmountFromCents(3_000_000_00)).
		AddCompensation("c1", "Tunjangan Transport", 300_000_00).
		AddDeduction("d1", "Potongan", 100_000_00).
		Build()

	want := int64(3_000_000_00 + 300_000_00 - 100_000_00)
	if ps.NetSalary.Cents() != want {
		t.Errorf("NetSalary = %d, want %d", ps.NetSalary.Cents(), want)
	}
}
