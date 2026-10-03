package usecase

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
)

type ManualPaySlipUsecase struct {
	periodRepo   repository.PayrollPeriodRepository
	paySlipRepo  repository.PaySlipRepository
	compItemRepo repository.CompensationItemRepository
	dedTypeRepo  repository.DeductionTypeRepository
	calcRepo     repository.PayrollCalculationRepository
}

func NewManualPaySlipUsecase(
	periodRepo repository.PayrollPeriodRepository,
	paySlipRepo repository.PaySlipRepository,
	compItemRepo repository.CompensationItemRepository,
	dedTypeRepo repository.DeductionTypeRepository,
	calcRepo repository.PayrollCalculationRepository,
) *ManualPaySlipUsecase {
	return &ManualPaySlipUsecase{
		periodRepo:   periodRepo,
		paySlipRepo:  paySlipRepo,
		compItemRepo: compItemRepo,
		dedTypeRepo:  dedTypeRepo,
		calcRepo:     calcRepo,
	}
}

func (uc *ManualPaySlipUsecase) CreateManualPaySlip(ctx context.Context, input models.ManualPaySlipInput) (*entity.PaySlip, error) {
	p, err := uc.periodRepo.FindByID(ctx, input.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("find period: %w", err)
	}
	if p == nil {
		return nil, errors.NewNotFound("period not found")
	}
	if p.Status == entity.PeriodStatusClosed {
		return nil, errors.NewInvalidInput("cannot create payslip for a closed period")
	}

	currency := entity.CurrencyFromDB(input.Currency)

	builder := entity.NewPaySlipBuilder(input.PeriodID, input.EmployeeID).
		WithCurrency(currency).
		WithSource(entity.PaySlipSourceManual)

	baseAmt, err := entity.NewAmount(input.BaseSalary)
	if err != nil {
		return nil, errors.NewInvalidInput("invalid base salary: " + err.Error())
	}
	builder.WithBaseSalary(baseAmt)

	for _, c := range input.Compensations {
		ci, err := uc.compItemRepo.FindByID(ctx, c.CompensationItemID)
		if err != nil {
			return nil, fmt.Errorf("find compensation item: %w", err)
		}
		if ci == nil {
			return nil, errors.NewNotFound("compensation item not found: " + c.CompensationItemID)
		}
		amt, err := entity.NewAmount(c.Amount)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid compensation amount: " + err.Error())
		}
		builder.AddCompensation(c.CompensationItemID, ci.Name, amt.Cents())
	}

	for _, d := range input.Deductions {
		dt, err := uc.dedTypeRepo.FindByID(ctx, d.DeductionTypeID)
		if err != nil {
			return nil, fmt.Errorf("find deduction type: %w", err)
		}
		if dt == nil {
			return nil, errors.NewNotFound("deduction type not found: " + d.DeductionTypeID)
		}
		amt, err := entity.NewAmount(d.Amount)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid deduction amount: " + err.Error())
		}
		builder.AddDeduction(d.DeductionTypeID, dt.Name, amt.Cents())
	}

	// Absent days are recorded on the slip for reporting only. The amount deducted for
	// absence comes from the per_day deduction type during processing, so there is no
	// absent_deduction amount to enter by hand (its column was dropped in migration 000039).
	builder.WithAbsentDays(input.AbsentDays)

	// Validated before the build so a bad figure aborts before anything is written.
	incomeInputs, err := entity.ValidateIncomeInputs(input.IncomeInputs)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	ps := builder.Build()
	ps.IncomeInputs = incomeInputs

	if err := uc.paySlipRepo.Upsert(ctx, ps); err != nil {
		return nil, fmt.Errorf("upsert manual payslip: %w", err)
	}

	// Upsert is INSERT ... ON CONFLICT (period_id, employee_id) DO UPDATE, and the DO
	// UPDATE list deliberately leaves id alone. So when the employee already had a slip in
	// this period the row kept its original primary key while ps still carries the one the
	// builder just generated. Returning ps would hand the caller an ID that does not exist,
	// and every later PUT against it would 404. Reading the row back returns the ID that is
	// actually stored.
	// Looked up by the pair that Upsert conflicts on, not by ps.ID: on a conflict that ID
	// is precisely the one that was not used.
	stored, err := uc.paySlipRepo.FindByEmployeeAndPeriod(ctx, input.EmployeeID, input.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("read back payslip: %w", err)
	}
	if stored == nil {
		return nil, errors.NewNotFound("pay slip not found after upsert")
	}
	return stored, nil
}

// UpdatePaySlip applies a partial correction to an existing payslip. Omitted fields are
// left as they are; only the keys present in the request change. Editing flips the slip
// to source = manual, which makes the processor preserve it on the next re-process.
//
// Corrections stay scoped to the slip's own period: this never touches salaries,
// employee_compensations, deduction_types or any other master data, so the following
// periods recalculate from the unchanged components.
func (uc *ManualPaySlipUsecase) UpdatePaySlip(ctx context.Context, input models.UpdatePaySlipInput) (*entity.PaySlip, error) {
	ps, err := uc.paySlipRepo.FindByID(ctx, input.PaySlipID)
	if err != nil {
		return nil, fmt.Errorf("find pay slip: %w", err)
	}
	if ps == nil {
		return nil, errors.NewNotFound("pay slip not found")
	}

	period, err := uc.periodRepo.FindByID(ctx, ps.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("find period: %w", err)
	}
	if period == nil {
		return nil, errors.NewNotFound("period not found")
	}
	if period.Status == entity.PeriodStatusClosed {
		return nil, errors.NewInvalidInput("cannot update payslip for a closed period")
	}

	if input.BaseSalary != nil {
		amt, err := entity.NewAmount(*input.BaseSalary)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid base salary: " + err.Error())
		}
		ps.BaseSalary = amt
	}

	if input.Compensations != nil {
		comps, err := uc.resolveCompensations(ctx, *input.Compensations)
		if err != nil {
			return nil, err
		}
		ps.CompensationsBreakdown = comps
	}

	if input.Deductions != nil {
		deds, err := uc.resolveDeductions(ctx, *input.Deductions)
		if err != nil {
			return nil, err
		}
		ps.DeductionsBreakdown = deds
	}

	if input.AbsentDays != nil {
		if *input.AbsentDays < 0 {
			return nil, errors.NewInvalidInput("absent_days must be >= 0")
		}
		ps.AbsentDays = *input.AbsentDays
	}

	// Independent of the money above: assigning to the slice cannot move a total, and it is
	// skipped by recalculate, which has no source for a hand-typed figure.
	if input.IncomeInputs != nil {
		incomeInputs, err := entity.ValidateIncomeInputs(*input.IncomeInputs)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		ps.IncomeInputs = incomeInputs
	}

	// recalculate is opt-in because the default behaviour is deliberately additive: a
	// correction touches only what it names. Rebuilding the breakdown can silently move
	// amounts an admin already reviewed, so it must be requested explicitly.
	if input.Recalculate {
		if err := uc.recalculateBreakdown(ctx, period, ps); err != nil {
			return nil, err
		}
	}

	ps.MarkManual()

	if err := uc.paySlipRepo.Update(ctx, ps); err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewNotFound("pay slip not found")
		}
		return nil, fmt.Errorf("update pay slip: %w", err)
	}
	return ps, nil
}

// recalculateBreakdown rebuilds the stored breakdown from the employee's master
// components and the period's attendance data.
//
// It reuses repository.BuildBreakdowns, the same helper the period processor uses, so a
// payslip re-derived here is identical to one produced by processing the period. The
// base salary supplied in the same request wins over the stored one, which is the whole
// point of the flag: raising the salary must move the percentage and daily-wage
// deductions that are derived from it.
func (uc *ManualPaySlipUsecase) recalculateBreakdown(
	ctx context.Context,
	period *entity.PayrollPeriod,
	ps *entity.PaySlip,
) error {
	if uc.calcRepo == nil {
		return errors.NewInvalidInput("recalculate is not available in this deployment")
	}

	workingDays, err := uc.calcRepo.QueryEmployeeWorkingDaysBatch(ctx, []string{ps.EmployeeID}, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query working days: %w", err)
	}
	attendance, err := uc.calcRepo.QueryAttendanceDayCounts(ctx, []string{ps.EmployeeID}, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query attendance day counts: %w", err)
	}

	days := attendance[ps.EmployeeID]
	calcCtx := entity.CalcContext{
		BaseSalaryCents:  ps.BaseSalary.Cents(),
		WorkingDays:      workingDays[ps.EmployeeID],
		UnpaidAbsentDays: days.UnpaidAbsent,
		AttendedDays:     days.Attended,
	}

	comps, err := uc.calcRepo.QueryEmployeeCompensations(ctx, ps.EmployeeID, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query compensations: %w", err)
	}
	deds, err := uc.calcRepo.QueryEmployeeDeductions(ctx, ps.EmployeeID, period.StartDate, period.EndDate)
	if err != nil {
		return fmt.Errorf("query deductions: %w", err)
	}

	compBreakdown, dedBreakdown := repository.BuildBreakdowns(comps, deds, calcCtx)
	ps.CompensationsBreakdown = compBreakdown
	ps.DeductionsBreakdown = dedBreakdown
	// absent_days is an output of the recalculation, not an input, so it is taken from
	// attendance rather than left holding a value that no longer matches the breakdown.
	ps.AbsentDays = days.UnpaidAbsent
	return nil
}

// resolveCompensations maps item IDs to their stored names so a payslip stays readable
// on the PDF even if the master item is renamed later.
func (uc *ManualPaySlipUsecase) resolveCompensations(ctx context.Context, items []models.ManualCompensationInput) ([]entity.CompensationBreakdown, error) {
	out := make([]entity.CompensationBreakdown, 0, len(items))
	for _, c := range items {
		ci, err := uc.compItemRepo.FindByID(ctx, c.CompensationItemID)
		if err != nil {
			return nil, fmt.Errorf("find compensation item: %w", err)
		}
		if ci == nil {
			return nil, errors.NewNotFound("compensation item not found: " + c.CompensationItemID)
		}
		amt, err := entity.NewAmount(c.Amount)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid compensation amount: " + err.Error())
		}
		out = append(out, entity.CompensationBreakdown{
			CompensationItemID: c.CompensationItemID,
			Name:               ci.Name,
			Amount:             amt.Float(),
		})
	}
	return out, nil
}

func (uc *ManualPaySlipUsecase) resolveDeductions(ctx context.Context, items []models.ManualDeductionInput) ([]entity.DeductionBreakdown, error) {
	out := make([]entity.DeductionBreakdown, 0, len(items))
	for _, d := range items {
		dt, err := uc.dedTypeRepo.FindByID(ctx, d.DeductionTypeID)
		if err != nil {
			return nil, fmt.Errorf("find deduction type: %w", err)
		}
		if dt == nil {
			return nil, errors.NewNotFound("deduction type not found: " + d.DeductionTypeID)
		}
		amt, err := entity.NewAmount(d.Amount)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid deduction amount: " + err.Error())
		}
		out = append(out, entity.DeductionBreakdown{
			DeductionTypeID: d.DeductionTypeID,
			Name:            dt.Name,
			Amount:          amt.Float(),
		})
	}
	return out, nil
}
