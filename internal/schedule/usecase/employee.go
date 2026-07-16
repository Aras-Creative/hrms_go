package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/schedule/entity"
	"hrms/internal/schedule/repository"
	errors "hrms/internal/pkg/apperror"
)

// EmployeeFetcher resolves user_id → employee_id for self-service endpoints.
type EmployeeFetcher interface {
	FindByUserID(ctx context.Context, userID string) (string, error)
	FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error)
}

type EmployeeExistsChecker interface {
	ExistsByID(ctx context.Context, id string) (bool, error)
	FindExistingIDs(ctx context.Context, ids []string) (map[string]bool, error)
}

type EmployeePatternUsecase struct {
	ewpRepo    repository.EmployeeWorkPatternRepository
	wpRepo     repository.WorkPatternRepository
	empChecker EmployeeExistsChecker
}

func NewEmployeePatternUsecase(ewpRepo repository.EmployeeWorkPatternRepository, wpRepo repository.WorkPatternRepository, empChecker EmployeeExistsChecker) *EmployeePatternUsecase {
	return &EmployeePatternUsecase{ewpRepo: ewpRepo, wpRepo: wpRepo, empChecker: empChecker}
}

func (uc *EmployeePatternUsecase) Assign(ctx context.Context, input AssignPatternInput) (*AssignResult, error) {
	wp, err := uc.wpRepo.FindByID(ctx, input.WorkPatternID)
	if err != nil {
		return nil, fmt.Errorf("failed to find work pattern: %w", err)
	}
	if wp == nil {
		return nil, errors.NewNotFound("work pattern not found")
	}

	validFrom, err := parseValidFrom(input.ValidFrom)
	if err != nil {
		return nil, err
	}

	existingMap, err := uc.empChecker.FindExistingIDs(ctx, input.EmployeeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to check employees: %w", err)
	}

	result := &AssignResult{}
	for _, empID := range input.EmployeeIDs {
		if !existingMap[empID] {
			result.Failed = append(result.Failed, AssignError{
				EmployeeID: empID,
				Error:      "employee not found",
			})
			continue
		}
		if _, err := uc.assignForEmployee(ctx, empID, input.WorkPatternID, validFrom, input.ValidTo); err != nil {
			result.Failed = append(result.Failed, AssignError{
				EmployeeID: empID,
				Error:      err.Error(),
			})
		} else {
			result.Succeeded = append(result.Succeeded, empID)
		}
	}
	return result, nil
}

func parseValidFrom(s string) (time.Time, error) {
	if s == "" {
		s = time.Now().Format("2006-01-02")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return t, errors.NewInvalidInput("invalid valid_from date, expected yyyy-mm-dd")
	}
	return t, nil
}

func (uc *EmployeePatternUsecase) assignForEmployee(ctx context.Context, employeeID, workPatternID string, validFrom time.Time, validTo *time.Time) (*entity.EmployeeWorkPattern, error) {
	current, err := uc.ewpRepo.FindActiveByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find current pattern: %w", err)
	}
	if current != nil {
		endDate := validFrom.AddDate(0, 0, -1)
		if err := uc.ewpRepo.DeactivateCurrent(ctx, employeeID, endDate); err != nil {
			return nil, fmt.Errorf("failed to deactivate current pattern: %w", err)
		}
	}

	ewp := entity.NewEmployeeWorkPattern(employeeID, workPatternID, validFrom, validTo)
	if err := uc.ewpRepo.Create(ctx, ewp); err != nil {
		return nil, fmt.Errorf("failed to assign work pattern: %w", err)
	}
	return ewp, nil
}

func (uc *EmployeePatternUsecase) GetActive(ctx context.Context, employeeID string) (*entity.EmployeeWorkPattern, error) {
	ewp, err := uc.ewpRepo.FindActiveByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active pattern: %w", err)
	}
	return ewp, nil
}

func (uc *EmployeePatternUsecase) GetHistory(ctx context.Context, employeeID string) ([]*entity.EmployeeWorkPattern, error) {
	list, err := uc.ewpRepo.FindHistoryByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pattern history: %w", err)
	}
	return list, nil
}
