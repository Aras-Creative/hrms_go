package usecase

import (
	"context"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/processor"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
)

type ProcessorUsecase struct {
	periodRepo repository.PayrollPeriodRepository
	proc       *processor.PayrollProcessor
}

func NewProcessorUsecase(periodRepo repository.PayrollPeriodRepository, proc *processor.PayrollProcessor) *ProcessorUsecase {
	return &ProcessorUsecase{periodRepo: periodRepo, proc: proc}
}

func (uc *ProcessorUsecase) ProcessPeriod(ctx context.Context, id string) error {
	p, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find period: %w", err)
	}
	if p == nil {
		return errors.NewNotFound("period not found")
	}

	if p.Status == entity.PeriodStatusClosed {
		return errors.NewInvalidInput("cannot process a closed period")
	}

	if p.Status == entity.PeriodStatusProcessed {
		if err := p.MarkDraft(); err != nil {
			return fmt.Errorf("reset to draft: %w", err)
		}
		if err := uc.periodRepo.Update(ctx, p); err != nil {
			return fmt.Errorf("update period to draft: %w", err)
		}
	}

	if err := uc.proc.ProcessPeriod(ctx, p); err != nil {
		return fmt.Errorf("process period: %w", err)
	}

	if err := p.MarkProcessed(); err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}
	return uc.periodRepo.Update(ctx, p)
}
