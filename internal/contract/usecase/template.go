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
	existing, err := uc.tmplRepo.FindByName(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("find template by name: %w", err)
	}
	if existing != nil {
		return nil, errors.NewAlreadyExists("template with this name already exists")
	}

	e, err := entity.NewContractTemplate(
		input.Name,
		input.ContractType,
		input.Description,
		input.Data,
		input.Templates,
	)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	if err := uc.tmplRepo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return e, nil
}

func (uc *ContractUsecase) GetTemplate(ctx context.Context, id string) (*entity.ContractTemplate, error) {
	e, err := uc.tmplRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find template: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("contract template not found")
	}
	return e, nil
}

func (uc *ContractUsecase) ListTemplates(ctx context.Context, input models.ListTemplateInput) (*models.ListTemplateResult, error) {
	entities, total, err := uc.tmplRepo.FindAll(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}

	return &models.ListTemplateResult{Entities: entities, Total: total}, nil
}

func (uc *ContractUsecase) UpdateTemplate(ctx context.Context, input models.UpdateTemplateInput) (*entity.ContractTemplate, error) {
	e, err := uc.tmplRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("find template for update: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("contract template not found")
	}

	expectedUpdatedAt := e.UpdatedAt

	if err := e.Update(
		input.Name,
		input.ContractType,
		input.Description,
		input.IsActive,
		input.Data,
		input.Templates,
	); err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	if err := uc.tmplRepo.Update(ctx, e, expectedUpdatedAt); err != nil {
		errStr := err.Error()
		if errStr == "template not found or modified by another user" {
			return nil, errors.NewNotFound("contract template not found or modified by another user")
		}
		return nil, fmt.Errorf("update template: %w", err)
	}
	return e, nil
}

func (uc *ContractUsecase) DeleteTemplate(ctx context.Context, id string) error {
	existing, err := uc.tmplRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find template for delete: %w", err)
	}
	if existing == nil {
		return errors.NewNotFound("contract template not found")
	}

	count, err := uc.tmplRepo.CountByTemplateID(ctx, id)
	if err != nil {
		return fmt.Errorf("count contracts for template: %w", err)
	}
	if count > 0 {
		return errors.NewInvalidInput("cannot delete template with existing contracts")
	}

	if err := uc.tmplRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}
