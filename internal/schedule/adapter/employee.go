package adapter

import (
	"context"
	"fmt"

	emplRepo "hrms/internal/employee/repository"
	scheduleUc "hrms/internal/schedule/usecase"
)

type EmployeeExistsAdapter struct {
	repo emplRepo.EmployeeRepository
}

func NewEmployeeExistsAdapter(repo emplRepo.EmployeeRepository) *EmployeeExistsAdapter {
	return &EmployeeExistsAdapter{repo: repo}
}

func (a *EmployeeExistsAdapter) ExistsByID(ctx context.Context, id string) (bool, error) {
	e, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	return e != nil, nil
}

func (a *EmployeeExistsAdapter) FindExistingIDs(ctx context.Context, ids []string) (map[string]bool, error) {
	if len(ids) == 0 {
		return map[string]bool{}, nil
	}
	employees, err := a.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(ids))
	for _, e := range employees {
		if e != nil {
			result[e.ID] = true
		}
	}
	return result, nil
}

type EmployeeFetcherAdapter struct {
	repo emplRepo.EmployeeRepository
}

func NewEmployeeFetcherAdapter(repo emplRepo.EmployeeRepository) *EmployeeFetcherAdapter {
	return &EmployeeFetcherAdapter{repo: repo}
}

func (a *EmployeeFetcherAdapter) FindByUserID(ctx context.Context, userID string) (string, error) {
	e, err := a.repo.FindByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("find employee by user id: %w", err)
	}
	if e == nil {
		return "", nil
	}
	return e.ID, nil
}

func (a *EmployeeFetcherAdapter) FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error) {
	e, err := a.repo.FindByID(ctx, employeeID)
	if err != nil {
		return "", fmt.Errorf("find employee by id: %w", err)
	}
	if e == nil || e.UserID == nil {
		return "", nil
	}
	return *e.UserID, nil
}

var _ scheduleUc.EmployeeFetcher = (*EmployeeFetcherAdapter)(nil)
var _ scheduleUc.EmployeeExistsChecker = (*EmployeeExistsAdapter)(nil)
