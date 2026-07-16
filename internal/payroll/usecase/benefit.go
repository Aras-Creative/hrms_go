package usecase

import (
	"context"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
)

type BenefitUsecase struct {
	benefitTypeRepo repository.BenefitTypeRepository
	empBenefitRepo  repository.EmployeeBenefitRepository
	employeeFetcher EmployeeFetcher
}

func NewBenefitUsecase(benefitTypeRepo repository.BenefitTypeRepository, empBenefitRepo repository.EmployeeBenefitRepository, employeeFetcher EmployeeFetcher) *BenefitUsecase {
	return &BenefitUsecase{benefitTypeRepo: benefitTypeRepo, empBenefitRepo: empBenefitRepo, employeeFetcher: employeeFetcher}
}

func (uc *BenefitUsecase) CreateType(ctx context.Context, input models.CreateBenefitTypeInput) (*entity.BenefitType, error) {
	emplCT, err := entity.ParseContributionType(input.EmployerContributionType)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	empCT, err := entity.ParseContributionType(input.EmployeeContributionType)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	bt := entity.NewBenefitType(input.Name, input.Description, emplCT, input.EmployerContributionValue, empCT, input.EmployeeContributionValue)
	if err := uc.benefitTypeRepo.Create(ctx, bt); err != nil {
		return nil, fmt.Errorf("create benefit type: %w", err)
	}
	return bt, nil
}

func (uc *BenefitUsecase) ListTypes(ctx context.Context, page, perPage int, isActive *bool) ([]*entity.BenefitType, int64, error) {
	return uc.benefitTypeRepo.FindAll(ctx, repository.BenefitTypeFilter{IsActive: isActive, Page: page, PerPage: perPage})
}

func (uc *BenefitUsecase) ListAssignmentsByEmployee(ctx context.Context, employeeID string) ([]*entity.EmployeeBenefit, error) {
	items, _, err := uc.empBenefitRepo.FindAll(ctx, repository.EmpBenefitFilter{EmployeeID: employeeID, Page: 1, PerPage: 100})
	return items, err
}

func (uc *BenefitUsecase) ListAssignments(ctx context.Context, filter repository.EmpBenefitFilter) ([]*entity.EmployeeBenefit, int64, error) {
	return uc.empBenefitRepo.FindAll(ctx, filter)
}

func (uc *BenefitUsecase) GetTypeOptions(ctx context.Context) ([]*entity.BenefitType, error) {
	active := true
	items, _, err := uc.benefitTypeRepo.FindAll(ctx, repository.BenefitTypeFilter{IsActive: &active, Page: 1, PerPage: 1000})
	if err != nil {
		return nil, fmt.Errorf("list benefit types: %w", err)
	}
	return items, nil
}
