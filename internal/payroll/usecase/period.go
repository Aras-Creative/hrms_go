package usecase

import (
	"context"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
)

type PeriodUsecase struct {
	periodRepo  repository.PayrollPeriodRepository
	paySlipRepo repository.PaySlipRepository
	empFetcher  EmployeeFetcher
}

func NewPeriodUsecase(periodRepo repository.PayrollPeriodRepository, paySlipRepo repository.PaySlipRepository, empFetcher EmployeeFetcher) *PeriodUsecase {
	return &PeriodUsecase{periodRepo: periodRepo, paySlipRepo: paySlipRepo, empFetcher: empFetcher}
}

func (uc *PeriodUsecase) CreatePeriod(ctx context.Context, input models.CreatePeriodInput) (*entity.PayrollPeriod, error) {
	if !input.StartDate.Before(input.EndDate) {
		return nil, errors.NewInvalidInput("start_date must be before end_date")
	}

	overlap, err := uc.periodRepo.FindByOverlap(ctx, input.StartDate, input.EndDate, "")
	if err != nil {
		return nil, fmt.Errorf("check overlap: %w", err)
	}
	if overlap != nil {
		return nil, errors.NewInvalidInput("period overlaps with existing period: " + overlap.Name)
	}

	p := entity.NewPayrollPeriod(input.Name, input.StartDate, input.EndDate)
	if err := uc.periodRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create period: %w", err)
	}
	return p, nil
}

func (uc *PeriodUsecase) GetPeriod(ctx context.Context, id string) (*entity.PayrollPeriod, error) {
	p, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find period: %w", err)
	}
	if p == nil {
		return nil, errors.NewNotFound("period not found")
	}
	return p, nil
}

func (uc *PeriodUsecase) ListPeriods(ctx context.Context, page, perPage int) ([]*entity.PayrollPeriod, int64, error) {
	return uc.periodRepo.FindAll(ctx, page, perPage)
}

func (uc *PeriodUsecase) ClosePeriod(ctx context.Context, id string) (*entity.PayrollPeriod, error) {
	p, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find period: %w", err)
	}
	if p == nil {
		return nil, errors.NewNotFound("period not found")
	}
	if err := p.MarkClosed(); err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	if err := uc.periodRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("close period: %w", err)
	}
	return p, nil
}

func (uc *PeriodUsecase) ListPaySlips(ctx context.Context, periodID string) ([]*entity.PaySlip, error) {
	return uc.paySlipRepo.FindByPeriodID(ctx, periodID)
}

func (uc *PeriodUsecase) GetPaySlip(ctx context.Context, id string) (*entity.PaySlip, error) {
	ps, err := uc.paySlipRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find pay slip: %w", err)
	}
	if ps == nil {
		return nil, errors.NewNotFound("pay slip not found")
	}
	return ps, nil
}

func (uc *PeriodUsecase) GetMyPayslips(ctx context.Context, userID string) ([]*entity.PaySlip, error) {
	employeeID, err := uc.empFetcher.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find employee by user: %w", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for user")
	}
	return uc.paySlipRepo.FindByEmployeeID(ctx, employeeID)
}

func (uc *PeriodUsecase) GetMyPaySlip(ctx context.Context, userID, periodID string) (*entity.PaySlip, error) {
	employeeID, err := uc.empFetcher.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find employee by user: %w", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for user")
	}
	ps, err := uc.paySlipRepo.FindByEmployeeAndPeriod(ctx, employeeID, periodID)
	if err != nil {
		return nil, fmt.Errorf("find pay slip: %w", err)
	}
	if ps == nil {
		return nil, errors.NewNotFound("pay slip not found")
	}
	return ps, nil
}
