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

func (p *PayrollProcessor) ProcessPeriod(ctx context.Context, period *entity.PayrollPeriod) error {
	if period.Status != entity.PeriodStatusDraft {
		return fmt.Errorf("period %s is not in draft status", period.ID)
	}

	salaries, err := p.calcRepo.QueryActiveSalaries(ctx, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query salaries: %w", err)
	}

	return p.processSalaries(ctx, period, salaries)
}

func (p *PayrollProcessor) ProcessEmployees(ctx context.Context, period *entity.PayrollPeriod, employeeIDs []string) error {
	if period.Status != entity.PeriodStatusDraft {
		return fmt.Errorf("period %s is not in draft status", period.ID)
	}

	salaries, err := p.calcRepo.QueryActiveSalariesByIDs(ctx, period.StartDate, period.EndDate, employeeIDs)
	if err != nil {
		return fmt.Errorf("query salaries: %w", err)
	}

	return p.processSalaries(ctx, period, salaries)
}

func (p *PayrollProcessor) processSalaries(
	ctx context.Context,
	period *entity.PayrollPeriod,
	salaries []repository.CalcSalaryRow,
) error {
	if len(salaries) == 0 {
		return nil
	}

	employeeIDs := make([]string, len(salaries))
	for i, s := range salaries {
		employeeIDs[i] = s.EmployeeID
	}
	workingDaysMap, err := p.calcRepo.QueryEmployeeWorkingDaysBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query working days: %w", err)
	}

	for _, s := range salaries {
		wd := workingDaysMap[s.EmployeeID]
		if wd <= 0 {
			wd = 20
		}
		if err := p.processEmployee(ctx, period, s.EmployeeID, s.Amount, s.Currency, wd); err != nil {
			return fmt.Errorf("process employee %s: %w", s.EmployeeID, err)
		}
	}
	return nil
}

func (p *PayrollProcessor) processEmployee(
	ctx context.Context,
	period *entity.PayrollPeriod,
	employeeID string,
	salaryCents int64,
	currency string,
	workingDays int,
) error {
	comps, err := p.calcRepo.QueryEmployeeCompensations(ctx, employeeID, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query compensations: %w", err)
	}

	dedItems, err := p.calcRepo.QueryEmployeeDeductions(ctx, employeeID, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query deductions: %w", err)
	}

	builder := entity.NewPaySlipBuilder(period.ID, employeeID).
		WithBaseSalary(entity.AmountFromCents(salaryCents)).
		WithCurrency(entity.CurrencyFromDB(currency))

	for _, c := range comps {
		builder.AddCompensation(c.ID, c.Name, c.Amount)
	}

	for _, d := range dedItems {
		dedCents := d.CalculateCents(salaryCents)
		builder.AddDeduction(d.ID, d.Name, dedCents)
	}

	return p.paySlipRepo.Upsert(ctx, builder.Build())
}
