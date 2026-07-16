package adapter

import (
	"context"

	emplRepo "hrms/internal/employee/repository"
	"hrms/internal/attendance/usecase"
)

type EmployeeFetcherAdapter struct {
	repo emplRepo.EmployeeRepository
}

func NewEmployeeFetcherAdapter(repo emplRepo.EmployeeRepository) *EmployeeFetcherAdapter {
	return &EmployeeFetcherAdapter{repo: repo}
}

func (a *EmployeeFetcherAdapter) FindByUserID(ctx context.Context, userID string) (string, string, error) {
	e, err := a.repo.FindByUserID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if e == nil {
		return "", "", nil
	}
	return e.ID, e.FullName, nil
}

func (a *EmployeeFetcherAdapter) ExistsByID(ctx context.Context, id string) (bool, error) {
	e, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	return e != nil, nil
}

func (a *EmployeeFetcherAdapter) FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error) {
	e, err := a.repo.FindByID(ctx, employeeID)
	if err != nil {
		return "", err
	}
	if e == nil || e.UserID == nil {
		return "", nil
	}
	return *e.UserID, nil
}

var _ usecase.EmployeeFetcher = (*EmployeeFetcherAdapter)(nil)
var _ usecase.EmployeeExistenceChecker = (*EmployeeFetcherAdapter)(nil)
