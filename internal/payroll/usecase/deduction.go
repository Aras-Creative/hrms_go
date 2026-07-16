package usecase

import (
	"context"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
)

type DeductionUsecase struct {
	deductionTypeRepo repository.DeductionTypeRepository
	empDeductionRepo  repository.EmployeeDeductionRepository
	employeeFetcher   EmployeeFetcher
}

func NewDeductionUsecase(deductionTypeRepo repository.DeductionTypeRepository, empDeductionRepo repository.EmployeeDeductionRepository, employeeFetcher EmployeeFetcher) *DeductionUsecase {
	return &DeductionUsecase{deductionTypeRepo: deductionTypeRepo, empDeductionRepo: empDeductionRepo, employeeFetcher: employeeFetcher}
}

func (uc *DeductionUsecase) CreateType(ctx context.Context, input models.CreateDeductionTypeInput) (*entity.DeductionType, error) {
	dtType, err := entity.ParseDeductionCalcType(input.DeductionType)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	dt := entity.NewDeductionType(input.Name, generateSlug(input.Name), input.Description, dtType, input.DefaultValue, input.IsMandatory)
	if err := uc.deductionTypeRepo.Create(ctx, dt); err != nil {
		return nil, fmt.Errorf("create deduction type: %w", err)
	}
	return dt, nil
}

func (uc *DeductionUsecase) ListTypes(ctx context.Context, page, perPage int, isActive, isMandatory *bool) ([]*entity.DeductionType, int64, error) {
	return uc.deductionTypeRepo.FindAll(ctx, repository.DeductionTypeFilter{IsActive: isActive, IsMandatory: isMandatory, Page: page, PerPage: perPage})
}

func (uc *DeductionUsecase) ListAssignmentsByEmployee(ctx context.Context, employeeID string) ([]*entity.EmployeeDeduction, error) {
	items, _, err := uc.empDeductionRepo.FindAll(ctx, repository.EmpDeductionFilter{EmployeeID: employeeID, Page: 1, PerPage: 100})
	return items, err
}

func (uc *DeductionUsecase) ListAssignments(ctx context.Context, filter repository.EmpDeductionFilter) ([]*entity.EmployeeDeduction, int64, error) {
	return uc.empDeductionRepo.FindAll(ctx, filter)
}

func (uc *DeductionUsecase) GetTypeOptions(ctx context.Context) ([]*entity.DeductionType, error) {
	active := true
	items, _, err := uc.deductionTypeRepo.FindAll(ctx, repository.DeductionTypeFilter{IsActive: &active, Page: 1, PerPage: 1000})
	if err != nil {
		return nil, fmt.Errorf("list deduction types: %w", err)
	}
	return items, nil
}
