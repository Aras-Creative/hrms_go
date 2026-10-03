package processor

import (
	"context"
	"testing"
	"time"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/repository"
)

// --- fakes ---

type fakeSlipRepo struct {
	// manual holds employee IDs that already carry a deliberate manual override.
	manual map[string]bool
	writes []string
}

func (r *fakeSlipRepo) Upsert(context.Context, *entity.PaySlip) error { return nil }

func (r *fakeSlipRepo) UpsertIfNotManual(_ context.Context, ps *entity.PaySlip) (bool, error) {
	if r.manual[ps.EmployeeID] {
		return true, nil
	}
	r.writes = append(r.writes, ps.EmployeeID)
	return false, nil
}

func (r *fakeSlipRepo) Update(context.Context, *entity.PaySlip) error { return nil }
func (r *fakeSlipRepo) FindByPeriodID(context.Context, string) ([]*entity.PaySlip, error) {
	return nil, nil
}
func (r *fakeSlipRepo) FindByID(context.Context, string) (*entity.PaySlip, error) {
	return nil, nil
}
func (r *fakeSlipRepo) FindByEmployeeID(context.Context, string) ([]*entity.PaySlip, error) {
	return nil, nil
}
func (r *fakeSlipRepo) FindByEmployeeAndPeriod(context.Context, string, string) (*entity.PaySlip, error) {
	return nil, nil
}
func (r *fakeSlipRepo) CountByPeriodID(context.Context, string) (int64, error) { return 0, nil }
func (r *fakeSlipRepo) DeleteByPeriodID(context.Context, string) error         { return nil }

type fakeCalcRepo struct {
	salaries []repository.CalcSalaryRow
}

func (r *fakeCalcRepo) QueryActiveSalaries(context.Context, time.Time, time.Time) ([]repository.CalcSalaryRow, error) {
	return r.salaries, nil
}

func (r *fakeCalcRepo) QueryActiveSalariesByIDs(context.Context, time.Time, time.Time, []string) ([]repository.CalcSalaryRow, error) {
	return r.salaries, nil
}

func (r *fakeCalcRepo) QueryAbsentDays(context.Context, string, time.Time, time.Time) (int, error) {
	return 0, nil
}

func (r *fakeCalcRepo) QueryAttendanceDayCounts(_ context.Context, _ []string, _, _ time.Time) (map[string]repository.PerDay, error) {
	return map[string]repository.PerDay{
		"emp-auto":   {UnpaidAbsent: 1, Attended: 19},
		"emp-manual": {UnpaidAbsent: 3, Attended: 17},
	}, nil
}

func (r *fakeCalcRepo) QueryEmployeeWorkingDaysBatch(context.Context, []string, time.Time, time.Time) (map[string]int, error) {
	return map[string]int{"emp-auto": 20, "emp-manual": 20}, nil
}

func (r *fakeCalcRepo) QueryEmployeeCompensations(context.Context, string, time.Time, time.Time) ([]repository.CalcCompRow, error) {
	return nil, nil
}

func (r *fakeCalcRepo) QueryEmployeeDeductions(context.Context, string, time.Time, time.Time) ([]repository.CalcDedRow, error) {
	return nil, nil
}

// --- tests ---

func newDraftPeriod() *entity.PayrollPeriod {
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	return entity.NewPayrollPeriod("Maret 2026", start, end)
}

func TestProcessPeriodPreservesManualPayslips(t *testing.T) {
	slipRepo := &fakeSlipRepo{manual: map[string]bool{"emp-manual": true}}
	calcRepo := &fakeCalcRepo{salaries: []repository.CalcSalaryRow{
		{EmployeeID: "emp-auto", Amount: 30_000_000_00, Currency: "IDR"},
		{EmployeeID: "emp-manual", Amount: 30_000_000_00, Currency: "IDR"},
	}}

	p := New(calcRepo, slipRepo)
	result, err := p.ProcessPeriod(context.Background(), newDraftPeriod())
	if err != nil {
		t.Fatalf("ProcessPeriod() error = %v", err)
	}

	if result.Generated != 1 {
		t.Errorf("Generated = %d, want 1", result.Generated)
	}
	if result.Skipped() != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped())
	}
	if len(result.SkippedEmployees) != 1 || result.SkippedEmployees[0] != "emp-manual" {
		t.Errorf("SkippedEmployees = %v, want [emp-manual]", result.SkippedEmployees)
	}

	// The manual slip must not be written at all, so its correction survives.
	if len(slipRepo.writes) != 1 || slipRepo.writes[0] != "emp-auto" {
		t.Errorf("writes = %v, want only [emp-auto]", slipRepo.writes)
	}
	for _, id := range slipRepo.writes {
		if id == "emp-manual" {
			t.Error("manual payslip was overwritten by the processor")
		}
	}
}

func TestProcessPeriodReportsNothingSkippedWhenAllAuto(t *testing.T) {
	slipRepo := &fakeSlipRepo{manual: map[string]bool{}}
	calcRepo := &fakeCalcRepo{salaries: []repository.CalcSalaryRow{
		{EmployeeID: "emp-a", Amount: 20_000_000_00, Currency: "IDR"},
		{EmployeeID: "emp-b", Amount: 25_000_000_00, Currency: "IDR"},
	}}

	p := New(calcRepo, slipRepo)
	result, err := p.ProcessPeriod(context.Background(), newDraftPeriod())
	if err != nil {
		t.Fatalf("ProcessPeriod() error = %v", err)
	}

	if result.Generated != 2 {
		t.Errorf("Generated = %d, want 2", result.Generated)
	}
	if result.Skipped() != 0 {
		t.Errorf("Skipped = %d, want 0", result.Skipped())
	}
}

func TestProcessPeriodRejectsNonDraftPeriod(t *testing.T) {
	slipRepo := &fakeSlipRepo{manual: map[string]bool{}}
	calcRepo := &fakeCalcRepo{}

	period := newDraftPeriod()
	if err := period.MarkProcessed(); err != nil {
		t.Fatalf("MarkProcessed() error = %v", err)
	}

	p := New(calcRepo, slipRepo)
	if _, err := p.ProcessPeriod(context.Background(), period); err == nil {
		t.Error("expected an error processing a non-draft period, got nil")
	}
}
