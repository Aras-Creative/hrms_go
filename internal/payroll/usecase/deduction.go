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

// resolvePerDaySource validates the value_source / unit_amount pair for the chosen
// basis. Only the per_day basis cares about the daily rate; anything else defaults
// to a fixed source with a zero unit amount.
func resolvePerDaySource(deductionType entity.DeductionCalcType, valueSource string, unitAmount float64) (entity.ValueSource, entity.Amount, error) {
	if deductionType != entity.DeductionCalcPerDay {
		return entity.ValueSourceFixed, entity.Amount{}, nil
	}

	if valueSource == "" {
		valueSource = string(entity.ValueSourceFixed)
	}
	vs, err := entity.ParseValueSource(valueSource)
	if err != nil {
		return "", entity.Amount{}, errors.NewInvalidInput(err.Error())
	}

	amount, err := entity.NewAmount(unitAmount)
	if err != nil {
		return "", entity.Amount{}, errors.NewInvalidInput("invalid unit_amount: " + err.Error())
	}
	if vs == entity.ValueSourceFixed && amount.Cents() <= 0 {
		return "", entity.Amount{}, errors.NewInvalidInput(
			"unit_amount must be greater than 0 when value_source is fixed; use daily_wage to derive the rate from the base salary",
		)
	}
	if vs == entity.ValueSourceDailyWage {
		amount = entity.Amount{}
	}
	return vs, amount, nil
}

func (uc *DeductionUsecase) CreateType(ctx context.Context, input models.CreateDeductionTypeInput) (*entity.DeductionType, error) {
	dtType, err := entity.ParseDeductionCalcType(input.DeductionType)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	valueSource, unitAmount, err := resolvePerDaySource(dtType, input.ValueSource, input.UnitAmount)
	if err != nil {
		return nil, err
	}

	dt := entity.NewDeductionType(input.Name, generateSlug(input.Name), input.Description, dtType, input.DefaultValue, valueSource, unitAmount, input.IsMandatory)
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

func (uc *DeductionUsecase) UpdateType(ctx context.Context, id string, input models.UpdateDeductionTypeInput) (*entity.DeductionType, error) {
	existing, err := uc.deductionTypeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.NewNotFound("deduction type not found")
	}
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Slug != nil {
		existing.Slug = *input.Slug
	}
	if input.DeductionType != nil {
		t, err := entity.ParseDeductionCalcType(*input.DeductionType)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		existing.DeductionType = t
	}
	if input.DefaultValue != nil {
		existing.DefaultValue = *input.DefaultValue
	}
	if input.ValueSource != nil {
		vs, err := entity.ParseValueSource(*input.ValueSource)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		existing.ValueSource = vs
	}
	if input.UnitAmount != nil {
		a, err := entity.NewAmount(*input.UnitAmount)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		existing.UnitAmount = a
	}
	if input.IsActive != nil {
		existing.IsActive = *input.IsActive
	}
	if input.IsMandatory != nil {
		existing.IsMandatory = *input.IsMandatory
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}
	if err := uc.deductionTypeRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (uc *DeductionUsecase) GetTypeByID(ctx context.Context, id string) (*entity.DeductionType, error) {
	return uc.deductionTypeRepo.FindByID(ctx, id)
}

func (uc *DeductionUsecase) DeleteType(ctx context.Context, id string) error {
	return uc.deductionTypeRepo.Delete(ctx, id)
}
