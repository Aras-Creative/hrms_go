package usecase

import (
	"context"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
)

type CompensationUsecase struct {
	compItemRepo   repository.CompensationItemRepository
	empCompRepo    repository.EmployeeCompensationRepository
	employeeFetcher EmployeeFetcher
}

func NewCompensationUsecase(compItemRepo repository.CompensationItemRepository, empCompRepo repository.EmployeeCompensationRepository, employeeFetcher EmployeeFetcher) *CompensationUsecase {
	return &CompensationUsecase{compItemRepo: compItemRepo, empCompRepo: empCompRepo, employeeFetcher: employeeFetcher}
}

func (uc *CompensationUsecase) CreateItem(ctx context.Context, input models.CreateCompensationItemInput) (*entity.CompensationItem, error) {
	itemType, err := entity.ParseCompensationItemType(input.ItemType)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	ci := entity.NewCompensationItem(input.Name, itemType, input.Description, input.IsTaxable)
	if err := uc.compItemRepo.Create(ctx, ci); err != nil {
		return nil, fmt.Errorf("create compensation item: %w", err)
	}
	return ci, nil
}

func (uc *CompensationUsecase) ListItems(ctx context.Context, page, perPage int, isActive *bool) ([]*entity.CompensationItem, int64, error) {
	return uc.compItemRepo.FindAll(ctx, repository.CompItemFilter{IsActive: isActive, Page: page, PerPage: perPage})
}

func (uc *CompensationUsecase) ListAssignmentsByEmployee(ctx context.Context, employeeID string) ([]*entity.EmployeeCompensation, error) {
	items, _, err := uc.empCompRepo.FindAll(ctx, repository.EmpCompFilter{EmployeeID: employeeID, Page: 1, PerPage: 100})
	return items, err
}

func (uc *CompensationUsecase) ListAssignments(ctx context.Context, filter repository.EmpCompFilter) ([]*entity.EmployeeCompensation, int64, error) {
	return uc.empCompRepo.FindAll(ctx, filter)
}

func (uc *CompensationUsecase) GetItemOptions(ctx context.Context) ([]*entity.CompensationItem, error) {
	active := true
	items, _, err := uc.compItemRepo.FindAll(ctx, repository.CompItemFilter{IsActive: &active, Page: 1, PerPage: 1000})
	if err != nil {
		return nil, fmt.Errorf("list compensation items: %w", err)
	}
	return items, nil
}
