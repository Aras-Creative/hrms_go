package processor

import (
	"context"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/repository"
)

type PayrollProcessor struct {
	calcRepo    repository.PayrollCalculationRepository
	paySlipRepo repository.PaySlipRepository
}

func New(calcRepo repository.PayrollCalculationRepository, paySlipRepo repository.PaySlipRepository) *PayrollProcessor {
	return &PayrollProcessor{
		calcRepo:    calcRepo,
		paySlipRepo: paySlipRepo,
	}
}

// Result reports what a processing run actually did. SkippedEmployees lists payslips
// that already carried a manual override and were therefore left untouched, so the
// caller can surface them instead of losing the correction silently.
type Result struct {
	Generated        int
	SkippedEmployees []string
}

func (r Result) Skipped() int { return len(r.SkippedEmployees) }

func (p *PayrollProcessor) ProcessPeriod(ctx context.Context, period *entity.PayrollPeriod) (Result, error) {
	if period.Status != entity.PeriodStatusDraft {
		return Result{}, fmt.Errorf("period %s is not in draft status", period.ID)
	}

	salaries, err := p.calcRepo.QueryActiveSalaries(ctx, period.StartDate, period.EndDate)
	if err != nil {
		return Result{}, fmt.Errorf("query salaries: %w", err)
	}

	return p.processSalaries(ctx, period, salaries)
}

func (p *PayrollProcessor) ProcessEmployees(ctx context.Context, period *entity.PayrollPeriod, employeeIDs []string) (Result, error) {
	if period.Status != entity.PeriodStatusDraft {
		return Result{}, fmt.Errorf("period %s is not in draft status", period.ID)
	}

	salaries, err := p.calcRepo.QueryActiveSalariesByIDs(ctx, period.StartDate, period.EndDate, employeeIDs)
	if err != nil {
		return Result{}, fmt.Errorf("query salaries: %w", err)
	}

	return p.processSalaries(ctx, period, salaries)
}

func (p *PayrollProcessor) processSalaries(
	ctx context.Context,
	period *entity.PayrollPeriod,
	salaries []repository.CalcSalaryRow,
) (Result, error) {
	var result Result
	if len(salaries) == 0 {
		return result, nil
	}

	employeeIDs := make([]string, len(salaries))
	for i, s := range salaries {
		employeeIDs[i] = s.EmployeeID
	}
	workingDaysMap, err := p.calcRepo.QueryEmployeeWorkingDaysBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return result, fmt.Errorf("query working days: %w", err)
	}
	attendanceMap, err := p.calcRepo.QueryAttendanceDayCounts(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return result, fmt.Errorf("query attendance day counts: %w", err)
	}

	for _, s := range salaries {
		days := attendanceMap[s.EmployeeID]
		calcCtx := entity.CalcContext{
			BaseSalaryCents:  s.Amount,
			WorkingDays:      workingDaysMap[s.EmployeeID],
			UnpaidAbsentDays: days.UnpaidAbsent,
			AttendedDays:     days.Attended,
		}
		skipped, err := p.processEmployee(ctx, period, s.EmployeeID, s.Currency, calcCtx)
		if err != nil {
			return result, fmt.Errorf("process employee %s: %w", s.EmployeeID, err)
		}
		if skipped {
			result.SkippedEmployees = append(result.SkippedEmployees, s.EmployeeID)
			continue
		}
		result.Generated++
	}
	return result, nil
}

func (p *PayrollProcessor) processEmployee(
	ctx context.Context,
	period *entity.PayrollPeriod,
	employeeID string,
	currency string,
	calcCtx entity.CalcContext,
) (bool, error) {
	comps, err := p.calcRepo.QueryEmployeeCompensations(ctx, employeeID, period.StartDate, period.EndDate)
	if err != nil {
		return false, fmt.Errorf("query compensations: %w", err)
	}

	dedItems, err := p.calcRepo.QueryEmployeeDeductions(ctx, employeeID, period.StartDate, period.EndDate)
	if err != nil {
		return false, fmt.Errorf("query deductions: %w", err)
	}

	compBreakdown, dedBreakdown := repository.BuildBreakdowns(comps, dedItems, calcCtx)

	builder := entity.NewPaySlipBuilder(period.ID, employeeID).
		WithBaseSalary(entity.AmountFromCents(calcCtx.BaseSalaryCents)).
		WithCurrency(entity.CurrencyFromDB(currency)).
		WithAbsentDays(calcCtx.UnpaidAbsentDays)

	for _, c := range compBreakdown {
		amount, err := entity.NewAmount(c.Amount)
		if err != nil {
			return false, fmt.Errorf("convert compensation amount: %w", err)
		}
		builder.AddCompensation(c.CompensationItemID, c.Name, amount.Cents())
	}
	for _, d := range dedBreakdown {
		amount, err := entity.NewAmount(d.Amount)
		if err != nil {
			return false, fmt.Errorf("convert deduction amount: %w", err)
		}
		builder.AddDeduction(d.DeductionTypeID, d.Name, amount.Cents())
	}

	// UpsertIfNotManual keeps an existing manual override intact and reports it as
	// skipped, so a re-process cannot discard a correction the admin entered by hand.
	return p.paySlipRepo.UpsertIfNotManual(ctx, builder.Build())
}
