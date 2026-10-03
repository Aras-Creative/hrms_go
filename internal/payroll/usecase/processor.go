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

// ProcessResult summarises a processing run for the API response.
type ProcessResult struct {
	Generated        int      `json:"generated"`
	Skipped          int      `json:"skipped"`
	SkippedEmployees []string `json:"skipped_employees,omitempty"`
}

// ProcessPeriod regenerates payslips for the period. Re-running on a processed period
// resets it to draft first. Payslips already marked source = manual are preserved and
// reported in the result rather than being recomputed away.
func (uc *ProcessorUsecase) ProcessPeriod(ctx context.Context, id string) (*ProcessResult, error) {
	p, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find period: %w", err)
	}
	if p == nil {
		return nil, errors.NewNotFound("period not found")
	}

	if p.Status == entity.PeriodStatusClosed {
		return nil, errors.NewInvalidInput("cannot process a closed period")
	}

	if p.Status == entity.PeriodStatusProcessed {
		if err := p.MarkDraft(); err != nil {
			return nil, fmt.Errorf("reset to draft: %w", err)
		}
		if err := uc.periodRepo.Update(ctx, p); err != nil {
			return nil, fmt.Errorf("update period to draft: %w", err)
		}
	}

	procResult, err := uc.proc.ProcessPeriod(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("process period: %w", err)
	}

	if err := p.MarkProcessed(); err != nil {
		return nil, fmt.Errorf("mark processed: %w", err)
	}
	if err := uc.periodRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("mark processed: %w", err)
	}

	return &ProcessResult{
		Generated:        procResult.Generated,
		Skipped:          procResult.Skipped(),
		SkippedEmployees: procResult.SkippedEmployees,
	}, nil
}
