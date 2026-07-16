package usecase

import (
	"context"
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
}

func NewManualPaySlipUsecase(
	periodRepo repository.PayrollPeriodRepository,
	paySlipRepo repository.PaySlipRepository,
	compItemRepo repository.CompensationItemRepository,
	dedTypeRepo repository.DeductionTypeRepository,
) *ManualPaySlipUsecase {
	return &ManualPaySlipUsecase{
		periodRepo:   periodRepo,
		paySlipRepo:  paySlipRepo,
		compItemRepo: compItemRepo,
		dedTypeRepo:  dedTypeRepo,
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

	if input.AbsentDays > 0 && input.AbsentDeduction > 0 {
		absentAmt, err := entity.NewAmount(input.AbsentDeduction)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid absent deduction: " + err.Error())
		}
		builder.WithAbsentDays(input.AbsentDays).
			AddDeduction("", "Absent Deduction", absentAmt.Cents())
	}

	ps := builder.Build()

	if err := uc.paySlipRepo.Upsert(ctx, ps); err != nil {
		return nil, fmt.Errorf("upsert manual payslip: %w", err)
	}
	return ps, nil
}
