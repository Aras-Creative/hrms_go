package adapter

import (
	"context"
	"time"

	scheduleEntity "hrms/internal/schedule/entity"
	scheduleRepo "hrms/internal/schedule/repository"
	contractUc "hrms/internal/contract/usecase"
)

type WorkPatternAssignerAdapter struct {
	wpRepo  scheduleRepo.WorkPatternRepository
	ewpRepo scheduleRepo.EmployeeWorkPatternRepository
}

func NewWorkPatternAssignerAdapter(wpRepo scheduleRepo.WorkPatternRepository, ewpRepo scheduleRepo.EmployeeWorkPatternRepository) *WorkPatternAssignerAdapter {
	return &WorkPatternAssignerAdapter{wpRepo: wpRepo, ewpRepo: ewpRepo}
}

func (a *WorkPatternAssignerAdapter) AssignWorkPattern(ctx context.Context, employeeID string, workPatternID *string, validFrom time.Time) error {
	var targetID string
	if workPatternID != nil && *workPatternID != "" {
		wp, err := a.wpRepo.FindByID(ctx, *workPatternID)
		if err != nil {
			return err
		}
		if wp == nil {
			patterns, err := a.wpRepo.FindAllActive(ctx)
			if err != nil || len(patterns) == 0 {
				return err
			}
			targetID = patterns[0].ID
		} else {
			targetID = wp.ID
		}
	} else {
		patterns, err := a.wpRepo.FindAllActive(ctx)
		if err != nil {
			return err
		}
		if len(patterns) == 0 {
			return nil
		}
		targetID = patterns[0].ID
	}

	existing, err := a.ewpRepo.FindActiveByEmployee(ctx, employeeID)
	if err != nil {
		return err
	}
	if existing != nil {
		endDate := validFrom.AddDate(0, 0, -1)
		if err := a.ewpRepo.DeactivateCurrent(ctx, employeeID, endDate); err != nil {
			return err
		}
	}

	ewp := scheduleEntity.NewEmployeeWorkPattern(employeeID, targetID, validFrom, nil)
	return a.ewpRepo.Upsert(ctx, ewp)
}

var _ contractUc.WorkPatternAssigner = (*WorkPatternAssignerAdapter)(nil)
