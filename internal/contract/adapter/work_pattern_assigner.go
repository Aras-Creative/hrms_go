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

func (a *WorkPatternAssignerAdapter) AssignDefaultWorkPattern(ctx context.Context, employeeID string, validFrom time.Time) error {
	patterns, err := a.wpRepo.FindAllActive(ctx)
	if err != nil {
		return err
	}
	if len(patterns) == 0 {
		return nil
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

	ewp := scheduleEntity.NewEmployeeWorkPattern(employeeID, patterns[0].ID, validFrom, nil)
	return a.ewpRepo.Upsert(ctx, ewp)
}

var _ contractUc.WorkPatternAssigner = (*WorkPatternAssignerAdapter)(nil)
