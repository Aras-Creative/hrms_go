package usecase

import (
	"context"
	"testing"
	"time"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
)

// stubCalcRepo embeds the interface so only the methods a test actually exercises need
// a body; anything else panics loudly instead of returning a quiet zero value.
type stubCalcRepo struct {
	repository.PayrollCalculationRepository
	comps    []repository.CalcCompRow
	deds     []repository.CalcDedRow
	days     map[string]repository.PerDay
	workDays map[string]int
}

func (s *stubCalcRepo) QueryEmployeeCompensations(context.Context, string) ([]repository.CalcCompRow, error) {
	return s.comps, nil
}
func (s *stubCalcRepo) QueryEmployeeDeductions(context.Context, string) ([]repository.CalcDedRow, error) {
	return s.deds, nil
}
func (s *stubCalcRepo) QueryAttendanceDayCounts(_ context.Context, ids []string, _, _ time.Time) (map[string]repository.PerDay, error) {
	out := map[string]repository.PerDay{}
	for _, id := range ids {
		out[id] = s.days[id]
	}
	return out, nil
}
func (s *stubCalcRepo) QueryEmployeeWorkingDaysBatch(_ context.Context, ids []string, _, _ time.Time) (map[string]int, error) {
	out := map[string]int{}
	for _, id := range ids {
		out[id] = s.workDays[id]
	}
	return out, nil
}

type stubPeriodRepo struct {
	repository.PayrollPeriodRepository
	period *entity.PayrollPeriod
}

func (s *stubPeriodRepo) FindByID(context.Context, string) (*entity.PayrollPeriod, error) {
	return s.period, nil
}

type stubPaySlipRepo struct {
	repository.PaySlipRepository
	slip    *entity.PaySlip
	updated *entity.PaySlip
}

func (s *stubPaySlipRepo) FindByID(context.Context, string) (*entity.PaySlip, error) {
	return s.slip, nil
}
func (s *stubPaySlipRepo) Update(_ context.Context, ps *entity.PaySlip) error {
	s.updated = ps
	return nil
}

func newRecalcFixture() (*ManualPaySlipUsecase, *stubPaySlipRepo) {
	const employeeID = "emp-1"

	period := &entity.PayrollPeriod{
		ID:        "period-1",
		StartDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
		Status:    entity.PeriodStatusProcessed,
	}

	base := entity.AmountFromCents(1_000_000_000) // Rp 10.000.000
	slip := entity.NewPaySlipBuilder("period-1", employeeID).
		WithBaseSalary(base).
		WithCurrency(entity.CurrencyFromDB("IDR")).
		WithAbsentDays(1).
		Build()
	// Stale figures from an earlier period, as if produced when the base was Rp 8 juta.
	slip.CompensationsBreakdown = []entity.CompensationBreakdown{{Name: "Lama", Amount: 111}}
	slip.DeductionsBreakdown = []entity.DeductionBreakdown{{Name: "Potongan Lama", Amount: 222}}
	slip.RecalculateTotals()

	calcRepo := &stubCalcRepo{
		comps: []repository.CalcCompRow{
			{ID: "c1", Name: "Tunjangan Transport", Amount: 5_000_000, CalcType: string(entity.CompensationCalcPerAttendedDay)},
		},
		deds: []repository.CalcDedRow{
			{ID: "d1", Name: "Potongan Alpa", Type: "per_day", ValueSource: "daily_wage"},
			{ID: "d2", Name: "Potongan Klaim", Type: "percentage", Value: 5},
			// Resolves to zero, so it must not appear on the payslip.
			{ID: "d3", Name: "Potongan Nol", Type: "percentage", Value: 0},
		},
		days:     map[string]repository.PerDay{employeeID: {UnpaidAbsent: 3, Attended: 19}},
		workDays: map[string]int{employeeID: 22},
	}

	slips := &stubPaySlipRepo{slip: slip}
	uc := NewManualPaySlipUsecase(
		&stubPeriodRepo{period: period},
		slips,
		nil,
		nil,
		calcRepo,
	)
	return uc, slips
}

func findDed(deds []entity.DeductionBreakdown, id string) (entity.DeductionBreakdown, bool) {
	for _, d := range deds {
		if d.DeductionTypeID == id {
			return d, true
		}
	}
	return entity.DeductionBreakdown{}, false
}

// TestUpdatePaySlipRecalculateRebuildsFromMaster is the regression test for the trap where
// raising base_salary left the percentage deduction pointing at the old salary.
func TestUpdatePaySlipRecalculateRebuildsFromMaster(t *testing.T) {
	uc, slips := newRecalcFixture()

	newBase := 12_000_000.0 // Rp 12.000.000
	got, err := uc.UpdatePaySlip(context.Background(), models.UpdatePaySlipInput{
		PaySlipID:   "slip-1",
		BaseSalary:  &newBase,
		Recalculate: true,
	})
	if err != nil {
		t.Fatalf("UpdatePaySlip() error = %v", err)
	}

	if slips.updated == nil {
		t.Fatal("expected the payslip to be persisted")
	}

	// 19 present days x Rp 50.000
	if len(got.CompensationsBreakdown) != 1 {
		t.Fatalf("compensations = %d lines, want 1", len(got.CompensationsBreakdown))
	}
	if c := got.CompensationsBreakdown[0]; c.Amount != 950_000 {
		t.Errorf("compensation amount = %v, want 950000", c.Amount)
	}

	// The percentage must be taken from the NEW base: 5% of Rp 12.000.000 = Rp 600.000.
	// The old value was Rp 500.000, derived from the previous Rp 10.000.000 salary.
	claim, ok := findDed(got.DeductionsBreakdown, "d2")
	if !ok {
		t.Fatal("percentage deduction missing from breakdown")
	}
	if claim.Amount != 600_000 {
		t.Errorf("percentage deduction = %v, want 600000 (5%% of the new base)", claim.Amount)
	}

	// 3 unpaid absent days x (Rp 12.000.000 / 22 working days)
	alpa, ok := findDed(got.DeductionsBreakdown, "d1")
	if !ok {
		t.Fatal("attendance deduction missing from breakdown")
	}
	if want := int64(3 * (1_200_000_000 / 22)); centsOf(t, alpa.Amount) != want {
		t.Errorf("attendance deduction = %d cents, want %d", centsOf(t, alpa.Amount), want)
	}

	if _, present := findDed(got.DeductionsBreakdown, "d3"); present {
		t.Error("zero-value deduction should have been dropped")
	}

	if got.AbsentDays != 3 {
		t.Errorf("absent_days = %d, want 3 (recomputed from attendance)", got.AbsentDays)
	}
	if got.Source != entity.PaySlipSourceManual {
		t.Errorf("source = %q, want %q", got.Source, entity.PaySlipSourceManual)
	}

	// Totals must agree with the rebuilt breakdown.
	// Compare in cents: major units are float64 and repeated addition drifts.
	var wantTotalComp, wantTotalDed int64
	for _, c := range got.CompensationsBreakdown {
		wantTotalComp += centsOf(t, c.Amount)
	}
	for _, d := range got.DeductionsBreakdown {
		wantTotalDed += centsOf(t, d.Amount)
	}
	if got.TotalCompensations.Cents() != wantTotalComp {
		t.Errorf("total_compensations = %d cents, want %d", got.TotalCompensations.Cents(), wantTotalComp)
	}
	if got.TotalDeductions.Cents() != wantTotalDed {
		t.Errorf("total_deductions = %d cents, want %d", got.TotalDeductions.Cents(), wantTotalDed)
	}
	wantNet := 1_200_000_000 + wantTotalComp - wantTotalDed
	if got.NetSalary.Cents() != wantNet {
		t.Errorf("net_salary = %d cents, want %d", got.NetSalary.Cents(), wantNet)
	}
}

// TestUpdatePaySlipWithoutRecalculateKeepsBreakdown pins the default behaviour: a salary
// change must not silently move amounts an admin already reviewed.
func TestUpdatePaySlipWithoutRecalculateKeepsBreakdown(t *testing.T) {
	uc, _ := newRecalcFixture()

	newBase := 12_000_000.0
	got, err := uc.UpdatePaySlip(context.Background(), models.UpdatePaySlipInput{
		PaySlipID:  "slip-1",
		BaseSalary: &newBase,
	})
	if err != nil {
		t.Fatalf("UpdatePaySlip() error = %v", err)
	}

	if got.CompensationsBreakdown[0].Name != "Lama" {
		t.Error("breakdown should be untouched without recalculate")
	}
	if got.AbsentDays != 1 {
		t.Errorf("absent_days = %d, want 1 (unchanged without recalculate)", got.AbsentDays)
	}
}

// TestUpdatePaySlipSetsBaseSalaryLabel covers the custom base salary label used by schemes
// such as CRM ("Gaji Pokok (Skema)").
func TestUpdatePaySlipSetsBaseSalaryLabel(t *testing.T) {
	uc, slips := newRecalcFixture()

	label := "Gaji Pokok (Skema)"
	got, err := uc.UpdatePaySlip(context.Background(), models.UpdatePaySlipInput{
		PaySlipID:       "slip-1",
		BaseSalaryLabel: &label,
	})
	if err != nil {
		t.Fatalf("UpdatePaySlip() error = %v", err)
	}
	if got.BaseSalaryLabel != label {
		t.Errorf("BaseSalaryLabel = %q, want %q", got.BaseSalaryLabel, label)
	}
	if slips.updated == nil || slips.updated.BaseSalaryLabel != label {
		t.Errorf("persisted label = %+v", slips.updated)
	}
}

// An explicit empty string must restore the default, not be skipped like an omitted key.
func TestUpdatePaySlipClearsBaseSalaryLabel(t *testing.T) {
	uc, _ := newRecalcFixture()

	empty := ""
	got, err := uc.UpdatePaySlip(context.Background(), models.UpdatePaySlipInput{
		PaySlipID:       "slip-1",
		BaseSalaryLabel: &empty,
	})
	if err != nil {
		t.Fatalf("UpdatePaySlip() error = %v", err)
	}
	if got.BaseSalaryLabel != "" {
		t.Errorf("BaseSalaryLabel = %q, want empty", got.BaseSalaryLabel)
	}
}

func TestUpdatePaySlipClosedPeriodRejected(t *testing.T) {
	uc, _ := newRecalcFixture()
	uc.periodRepo = &stubPeriodRepo{period: &entity.PayrollPeriod{
		ID:        "period-1",
		Status:    entity.PeriodStatusClosed,
		StartDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
	}}

	_, err := uc.UpdatePaySlip(context.Background(), models.UpdatePaySlipInput{
		PaySlipID:   "slip-1",
		Recalculate: true,
	})
	if err == nil {
		t.Fatal("expected recalculate on a closed period to be rejected")
	}
}

// centsOf converts a major-unit breakdown amount back to cents for exact comparison.
func centsOf(t *testing.T, major float64) int64 {
	t.Helper()
	amt, err := entity.NewAmount(major)
	if err != nil {
		t.Fatalf("breakdown amount %v is not a valid amount: %v", major, err)
	}
	return amt.Cents()
}
