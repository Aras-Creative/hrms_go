package repository

import (
	"testing"

	"hrms/internal/payroll/entity"
)

func TestCalcCompRowCalculateCents(t *testing.T) {
	const rate = 15_000_00 // Rp 15.000 per attended day

	tests := []struct {
		name string
		row  CalcCompRow
		ctx  entity.CalcContext
		want int64
	}{
		{
			name: "monthly allowance pays once",
			row:  CalcCompRow{Amount: 500_000_00, CalcType: string(entity.CompensationCalcFixed)},
			ctx:  entity.CalcContext{AttendedDays: 20},
			want: 500_000_00,
		},
		{
			name: "per_attended_day multiplies by days present",
			row:  CalcCompRow{Amount: rate, CalcType: string(entity.CompensationCalcPerAttendedDay)},
			ctx:  entity.CalcContext{AttendedDays: 18},
			want: rate * 18,
		},
		{
			name: "per_attended_day ignores absence and lateness",
			row:  CalcCompRow{Amount: rate, CalcType: string(entity.CompensationCalcPerAttendedDay)},
			ctx:  entity.CalcContext{AttendedDays: 18, UnpaidAbsentDays: 2, WorkingDays: 22},
			want: rate * 18,
		},
		{
			name: "per_attended_day with no present day pays nothing",
			row:  CalcCompRow{Amount: rate, CalcType: string(entity.CompensationCalcPerAttendedDay)},
			ctx:  entity.CalcContext{AttendedDays: 0, UnpaidAbsentDays: 22},
			want: 0,
		},
		{
			name: "per_attended_day counts attended days from present plus eligible leave (cuti/sick)",
			row:  CalcCompRow{Amount: rate, CalcType: string(entity.CompensationCalcPerAttendedDay)},
			ctx:  entity.CalcContext{AttendedDays: 17, UnpaidAbsentDays: 0, WorkingDays: 20},
			want: rate * 17,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.row.CalculateCents(tc.ctx); got != tc.want {
				t.Errorf("CalculateCents() = %d, want %d", got, tc.want)
			}
		})
	}
}
func TestCalcDedRowCalculateCents(t *testing.T) {
	const salary = 30_000_000_00

	tests := []struct {
		name string
		row  CalcDedRow
		ctx  entity.CalcContext
		want int64
	}{
		{
			name: "percentage of base salary",
			row:  CalcDedRow{Type: string(entity.DeductionCalcPercentage), Value: 5},
			ctx:  entity.CalcContext{BaseSalaryCents: salary},
			want: salary * 5 / 100,
		},
		{
			name: "fixed flat amount",
			row:  CalcDedRow{Type: string(entity.DeductionCalcFixed), Value: 500_000},
			ctx:  entity.CalcContext{BaseSalaryCents: salary},
			want: 500_000_00,
		},
		{
			name: "per_day fixed source",
			row: CalcDedRow{
				Type:        string(entity.DeductionCalcPerDay),
				ValueSource: string(entity.ValueSourceFixed),
				UnitAmount:  25_000_00,
			},
			ctx:  entity.CalcContext{BaseSalaryCents: salary, UnpaidAbsentDays: 3},
			want: 75_000_00,
		},
		{
			name: "per_day daily_wage source",
			row: CalcDedRow{
				Type:        string(entity.DeductionCalcPerDay),
				ValueSource: string(entity.ValueSourceDailyWage),
			},
			ctx:  entity.CalcContext{BaseSalaryCents: salary, WorkingDays: 20, UnpaidAbsentDays: 4},
			want: (salary / 20) * 4,
		},
		{
			// Regression: the override used to be discarded whenever the master said
			// daily_wage, so the employee row validated and persisted but never paid.
			name: "per_day daily_wage master loses to an explicit override",
			row: CalcDedRow{
				Type:                 string(entity.DeductionCalcPerDay),
				ValueSource:          string(entity.ValueSourceDailyWage),
				UnitAmount:           15_000_00,
				UnitAmountOverridden: true,
			},
			ctx:  entity.CalcContext{BaseSalaryCents: salary, WorkingDays: 20, UnpaidAbsentDays: 4},
			want: 60_000_00,
		},
		{
			// A zero override is still an override: 0 is how a per_day type is
			// explicitly disabled for one employee.
			name: "per_day daily_wage master with an explicit zero override pays nothing",
			row: CalcDedRow{
				Type:                 string(entity.DeductionCalcPerDay),
				ValueSource:          string(entity.ValueSourceDailyWage),
				UnitAmountOverridden: true,
			},
			ctx:  entity.CalcContext{BaseSalaryCents: salary, WorkingDays: 20, UnpaidAbsentDays: 4},
			want: 0,
		},
		{
			name: "per_day fixed master with a zero override pays nothing",
			row: CalcDedRow{
				Type:                 string(entity.DeductionCalcPerDay),
				ValueSource:          string(entity.ValueSourceFixed),
				UnitAmountOverridden: true,
			},
			ctx:  entity.CalcContext{BaseSalaryCents: salary, UnpaidAbsentDays: 4},
			want: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.row.CalculateCents(tc.ctx); got != tc.want {
				t.Errorf("CalculateCents() = %d, want %d", got, tc.want)
			}
		})
	}
}

// The payslip and the period overview must never disagree on an amount.
func TestDeductionRowMatchesTypeCalculate(t *testing.T) {
	ctxs := []entity.CalcContext{
		{BaseSalaryCents: 30_000_000_00, WorkingDays: 22, UnpaidAbsentDays: 3, AttendedDays: 19},
		{BaseSalaryCents: 7_500_000_00, WorkingDays: 20, UnpaidAbsentDays: 0, AttendedDays: 20},
		{BaseSalaryCents: 45_000_000_00, WorkingDays: 0, UnpaidAbsentDays: 11, AttendedDays: 9},
	}

	types := []entity.DeductionType{
		{DeductionType: entity.DeductionCalcPercentage, DefaultValue: 2.5},
		{DeductionType: entity.DeductionCalcFixed, DefaultValue: 750_000},
		{DeductionType: entity.DeductionCalcPerDay, ValueSource: entity.ValueSourceFixed, UnitAmount: entity.AmountFromCents(30_000_00)},
		{DeductionType: entity.DeductionCalcPerDay, ValueSource: entity.ValueSourceDailyWage},
	}

	for _, dt := range types {
		for _, ctx := range ctxs {
			row := CalcDedRow{
				ID:          dt.ID,
				Name:        dt.Name,
				Type:        string(dt.DeductionType),
				Value:       dt.DefaultValue,
				ValueSource: string(dt.ValueSource),
				UnitAmount:  dt.UnitAmount.Cents(),
			}
			if got, want := row.CalculateCents(ctx), dt.Calculate(ctx); got != want {
				t.Errorf("basis=%s source=%s ctx=%+v: row=%d type=%d", dt.DeductionType, dt.ValueSource, ctx, got, want)
			}
			if got, want := row.PerDayRateCents(ctx), dt.PerDayRateCents(ctx); got != want {
				t.Errorf("basis=%s source=%s: perDayRate row=%d type=%d", dt.DeductionType, dt.ValueSource, got, want)
			}
		}
	}
}

// The payroll processor and the period overview must agree on the clamp. This row is read
// straight from employee_deductions, so it is the path that actually produced the reported
// negative payslip; clamping only the entity version would leave payroll itself unbounded.
func TestCalcDedRowPercentageClamped(t *testing.T) {
	base := int64(235_000_000) // Rp 2.350.000

	legacy := CalcDedRow{Type: "percentage", Value: 100000}
	got := legacy.CalculateCents(entity.CalcContext{BaseSalaryCents: base})
	if got != base {
		t.Errorf("CalculateCents() = %d, want %d: a 100000%% deduction must not exceed the salary", got, base)
	}

	sane := CalcDedRow{Type: "percentage", Value: 4}
	if got := sane.CalculateCents(entity.CalcContext{BaseSalaryCents: base}); got != 9_400_000 {
		t.Errorf("4%% of %d = %d, want 9400000", base, got)
	}
}
