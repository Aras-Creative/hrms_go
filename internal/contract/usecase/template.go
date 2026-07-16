package usecase

import (
	"context"
	"fmt"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
	"hrms/internal/contract/repository"
	"hrms/internal/employee/numbergen"
	errors "hrms/internal/pkg/apperror"
)

type EmployeeFetcher interface {
	FindEmployeeRenderData(ctx context.Context, id string) (*entity.EmployeeRenderData, error)
	FindDesignationIDs(ctx context.Context, ids []string) (map[string]*string, error)
	FindEmployeeIDByUserID(ctx context.Context, userID string) (string, error)
	FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error)
	FindBriefByIDs(ctx context.Context, ids []string) (map[string]EmployeeBrief, error)
}

type EmployeeBrief struct {
	Name           string
	ProfilePhotoURL string
}

type DesignationFetcher interface {
	FindNamesByIDs(ctx context.Context, ids []string) (map[string]string, error)
}

type SalaryFetcher interface {
	FindCurrentByEmployeeIDs(ctx context.Context, employeeIDs []string) (map[string]string, error)
}

type ContractUsecase struct {
	tmplRepo     repository.TemplateRepository
	contractRepo repository.ContractRepository
	signingRepo  repository.SigningRepository
	numGen       *numbergen.Generator
	empFetcher   EmployeeFetcher
	desFetcher   DesignationFetcher
	salFetcher   SalaryFetcher
}

func NewContractUsecase(
	tmplRepo repository.TemplateRepository,
	contractRepo repository.ContractRepository,
	signingRepo repository.SigningRepository,
	numGen *numbergen.Generator,
	empFetcher EmployeeFetcher,
	desFetcher DesignationFetcher,
	salFetcher SalaryFetcher,
) *ContractUsecase {
	return &ContractUsecase{
		tmplRepo:     tmplRepo,
		contractRepo: contractRepo,
		signingRepo:  signingRepo,
		numGen:       numGen,
		empFetcher:   empFetcher,
		desFetcher:   desFetcher,
		salFetcher:   salFetcher,
	}
}

func (uc *ContractUsecase) CreateTemplate(ctx context.Context, input models.CreateTemplateInput) (*entity.ContractTemplate, error) {
	contractType, err := entity.ParseContractType(input.ContractType)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	e := entity.NewContractTemplate(
		input.Name,
		contractType,
		input.Description,
		input.Data,
		input.Templates,
	)
	if err := uc.tmplRepo.Create(ctx, e); err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to create template: %v", err))
	}
	return e, nil
}

func (uc *ContractUsecase) GetTemplate(ctx context.Context, id string) (*entity.ContractTemplate, error) {
	e, err := uc.tmplRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find template: %v", err))
	}
	if e == nil {
		return nil, errors.NewNotFound("contract template not found")
	}
	return e, nil
}

func (uc *ContractUsecase) ListTemplates(ctx context.Context, input models.ListTemplateInput) (*models.ListTemplateResult, error) {
	entities, total, err := uc.tmplRepo.FindAll(ctx, input)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to list templates: %v", err))
	}

	return &models.ListTemplateResult{Entities: entities, Total: total}, nil
}

func (uc *ContractUsecase) UpdateTemplate(ctx context.Context, input models.UpdateTemplateInput) (*entity.ContractTemplate, error) {
	e, err := uc.tmplRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find template for update: %v", err))
	}
	if e == nil {
		return nil, errors.NewNotFound("contract template not found")
	}

	contractType, err := entity.ParseContractType(input.ContractType)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	e.Update(
		input.Name,
		contractType,
		input.Description,
		input.IsActive,
		input.Data,
		input.Templates,
	)

	if err := uc.tmplRepo.Update(ctx, e); err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to update template: %v", err))
	}
	return e, nil
}

func (uc *ContractUsecase) DeleteTemplate(ctx context.Context, id string) error {
	if err := uc.tmplRepo.Delete(ctx, id); err != nil {
		return errors.NewInternal(fmt.Sprintf("failed to delete template: %v", err))
	}
	return nil
}
