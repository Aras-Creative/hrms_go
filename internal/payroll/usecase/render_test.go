package usecase

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
)

// A count printed through the money formatter would read "Rp 10" for "10 closings", which
// is the exact misreading the separate notes block exists to prevent.
func TestFormatIncomeInputValue(t *testing.T) {
	cases := []struct {
		in   entity.IncomeInput
		want string
	}{
		{entity.IncomeInput{Value: 10, Unit: entity.IncomeInputUnitNumber}, "10"},
		{entity.IncomeInput{Value: 12345.6, Unit: entity.IncomeInputUnitNumber}, "12345.6"},
		{entity.IncomeInput{Value: 5, Unit: entity.IncomeInputUnitPercent}, "5%"},
		{entity.IncomeInput{Value: 8, Unit: entity.IncomeInputUnitDays}, "8 hari"},
		{entity.IncomeInput{Value: 8}, "8"}, // empty unit defaults to a plain number
	}
	for _, c := range cases {
		if got := formatIncomeInputValue(c.in); got != c.want {
			t.Errorf("formatIncomeInputValue(%+v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIncomeInputLabel(t *testing.T) {
	if got := entity.IncomeInputLabel("persentase_rts"); got != "Persentase RTS" {
		t.Errorf("label = %q", got)
	}
	// Unknown keys must still print something identifying rather than a blank line.
	if got := entity.IncomeInputLabel("whatever"); got != "whatever" {
		t.Errorf("fallback = %q, want the key itself", got)
	}
}

// Exercises the embedded template without a browser. PrintPayslip needs Chrome to make a
// PDF, which is absent in CI, so this is what proves the notes section actually reaches the
// HTML and that the totals are untouched by it.
func TestPayslipTemplateRendersIncomeNotesOutsideTotals(t *testing.T) {
	inputs, err := entity.ValidateIncomeInputs([]entity.IncomeInput{
		{Key: "jumlah_sukses", Value: 10, Unit: entity.IncomeInputUnitNumber},
		{Key: "persentase_rts", Value: 5, Unit: entity.IncomeInputUnitPercent},
		{Key: "closing_bersih", Value: 8, Unit: entity.IncomeInputUnitNumber, Notes: "8 dari 10"},
	})
	if err != nil {
		t.Fatal(err)
	}

	ps := entity.ReconstitutePaySlip(
		"ps-1", "p-1", "e-1",
		5_000_000_00, 200_000_00, 100_000_00,
		4,
		5_100_000_00,
		"IDR", string(entity.PaySlipSourceManual),
		[]byte(`[{"compensation_item_id":"c1","name":"Bonus Closing Bersih","amount":200000}]`),
		[]byte(`[{"deduction_type_id":"d1","name":"BPJS","amount":100000}]`),
		[]byte(`[]`), time.Now(), time.Now(),
	)
	ps.IncomeInputs = inputs

	uc := &RenderUsecase{}
	data := uc.buildRenderData(ps,
		&entity.PayrollPeriod{Name: "Okt 2026"},
		models.PayslipEmployeeData{FullName: "Budi", EmployeeNumber: "EMP-001"},
		"PT Contoh", "Jl. Contoh", "")

	if data.NetSalary != "Rp 5,100,000" {
		t.Errorf("NetSalary = %q", data.NetSalary)
	}
	if len(data.IncomeNotes) != 3 {
		t.Fatalf("IncomeNotes = %+v", data.IncomeNotes)
	}

	tpl, err := template.New("p").Parse(payslipTemplate)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	html := buf.String()

	for _, want := range []string{
		"Catatan Perhitungan",
		"Jumlah Sukses", "10",
		"Persentase RTS", "5%",
		"Closing Bersih", "8 dari 10",
		"tidak termasuk dalam perhitungan gaji bersih",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("PDF HTML missing %q", want)
		}
	}

	// The notes block must come after the net pay, not inside the income table.
	if strings.Index(html, "Catatan Perhitungan") < strings.Index(html, "Gaji Bersih Diterima") {
		t.Error("notes section renders above the net pay, where they read as part of the sum")
	}
	// A count must never reach the page wearing a currency symbol.
	if strings.Contains(html, "Rp 10 ") || strings.Contains(html, "Rp 5%") {
		t.Error("a count or percentage was formatted as money")
	}
}

// Without figures the section is omitted entirely rather than printing an empty box.
func TestPayslipTemplateOmitsIncomeNotesWhenEmpty(t *testing.T) {
	ps := entity.ReconstitutePaySlip("ps-1", "p-1", "e-1", 0, 0, 0, 0, 0,
		"IDR", string(entity.PaySlipSourceAuto), nil, nil, []byte(`[]`), time.Now(), time.Now())
	uc := &RenderUsecase{}
	data := uc.buildRenderData(ps, &entity.PayrollPeriod{Name: "Okt"},
		models.PayslipEmployeeData{}, "", "", "")
	tpl, _ := template.New("p").Parse(payslipTemplate)
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "Catatan Perhitungan") {
		t.Error("empty notes should render no section at all")
	}
}
