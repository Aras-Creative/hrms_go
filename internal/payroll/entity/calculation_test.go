package entity

import (
	"testing"
	"time"
)

func TestCalcContextDailyRate(t *testing.T) {
	tests := []struct {
		name         string
		ctx          CalcContext
		wantDailyDay int
		wantRate     int64
	}{
		{
			name:         "uses scheduled working days",
			ctx:          CalcContext{BaseSalaryCents: 30_000_000_00, WorkingDays: 22},
			wantDailyDay: 22,
			wantRate:     30_000_000_00 / 22,
		},
		{
			name:         "falls back to default when schedule has no days",
			ctx:          CalcContext{BaseSalaryCents: 20_000_000_00, WorkingDays: 0},
			wantDailyDay: DefaultWorkingDays,
			wantRate:     20_000_000_00 / DefaultWorkingDays,
		},
		{
			name:         "falls back on negative schedule",
			ctx:          CalcContext{BaseSalaryCents: 20_000_000_00, WorkingDays: -3},
			wantDailyDay: DefaultWorkingDays,
			wantRate:     20_000_000_00 / DefaultWorkingDays,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ctx.EffectiveWorkingDays(); got != tc.wantDailyDay {
				t.Errorf("EffectiveWorkingDays() = %d, want %d", got, tc.wantDailyDay)
			}
			if got := tc.ctx.DailyRate(); got != tc.wantRate {
				t.Errorf("DailyRate() = %d, want %d", got, tc.wantRate)
			}
		})
	}
}

func TestDeductionTypeCalculate(t *testing.T) {
	const salary = 30_000_000_00 // Rp 30.000.000

	tests := []struct {
		name string
		dt   DeductionType
		ctx  CalcContext
		want int64
	}{
		{
			name: "percentage of base salary",
			dt:   DeductionType{DeductionType: DeductionCalcPercentage, DefaultValue: 5},
			ctx:  CalcContext{BaseSalaryCents: salary},
			want: salary * 5 / 100,
		},
		{
			name: "fixed flat amount for the period",
			dt:   DeductionType{DeductionType: DeductionCalcFixed, DefaultValue: 500_000},
			ctx:  CalcContext{BaseSalaryCents: salary, UnpaidAbsentDays: 5},
			want: 500_000_00,
		},
		{
			name: "per_day with fixed source uses unit amount",
			dt: DeductionType{
				DeductionType: DeductionCalcPerDay,
				ValueSource:   ValueSourceFixed,
				UnitAmount:    AmountFromCents(25_000_00),
			},
			ctx:  CalcContext{BaseSalaryCents: salary, UnpaidAbsentDays: 3},
			want: 75_000_00,
		},
		{
			name: "per_day with daily_wage source derives from salary",
			dt: DeductionType{
				DeductionType: DeductionCalcPerDay,
				ValueSource:   ValueSourceDailyWage,
			},
			ctx:  CalcContext{BaseSalaryCents: salary, WorkingDays: 20, UnpaidAbsentDays: 4},
			want: (salary / 20) * 4,
		},
		{
			name: "per_day ignores unit amount when source is daily_wage",
			dt: DeductionType{
				DeductionType: DeductionCalcPerDay,
				ValueSource:   ValueSourceDailyWage,
				UnitAmount:    AmountFromCents(999_999_99),
			},
			ctx:  CalcContext{BaseSalaryCents: salary, WorkingDays: 20, UnpaidAbsentDays: 2},
			want: (salary / 20) * 2,
		},
		{
			name: "per_day with zero absent days is free",
			dt: DeductionType{
				DeductionType: DeductionCalcPerDay,
				ValueSource:   ValueSourceFixed,
				UnitAmount:    AmountFromCents(25_000_00),
			},
			ctx:  CalcContext{BaseSalaryCents: salary, UnpaidAbsentDays: 0},
			want: 0,
		},
		{
			name: "per_day daily_wage guards against missing work schedule",
			dt: DeductionType{
				DeductionType: DeductionCalcPerDay,
				ValueSource:   ValueSourceDailyWage,
			},
			ctx:  CalcContext{BaseSalaryCents: salary, WorkingDays: 0, UnpaidAbsentDays: 2},
			want: (salary / DefaultWorkingDays) * 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.dt.Calculate(tc.ctx); got != tc.want {
				t.Errorf("Calculate() = %d, want %d", got, tc.want)
			}
		})
	}
}

func testTime() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
