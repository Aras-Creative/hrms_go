package usecase

import (
	"context"
	"fmt"

	"hrms/internal/designation/entity"
	"hrms/internal/designation/models"
	"hrms/internal/designation/repository"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
)

type DesignationUsecase struct {
	repo repository.DesignationRepository
}

func NewDesignationUsecase(repo repository.DesignationRepository) *DesignationUsecase {
	return &DesignationUsecase{repo: repo}
}

func (uc *DesignationUsecase) Create(ctx context.Context, name string) (*entity.Designation, error) {
	code := uc.AcronymFromName(name)
	d := entity.NewDesignation(name, code)
	if err := uc.repo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to create designation: %w", err)
	}
	return d, nil
}

func (uc *DesignationUsecase) FindByID(ctx context.Context, id string) (*entity.Designation, error) {
	d, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find designation: %w", err)
	}
	if d == nil {
		return nil, errors.NewNotFound("designation not found")
	}
	return d, nil
}

func (uc *DesignationUsecase) FindAll(ctx context.Context) ([]models.DesignationReadModel, error) {
	list, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list designations: %w", err)
	}
	return list, nil
}

func (uc *DesignationUsecase) Options(ctx context.Context) []response.Option {
	list, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil
	}
	opts := make([]response.Option, 0, len(list))
	for _, d := range list {
		opts = append(opts, response.Option{Value: d.ID, Label: fmt.Sprintf("%s (%s)", d.Name, d.Code), Extra: d.Code})
	}
	return opts
}

func (uc *DesignationUsecase) Update(ctx context.Context, id, name string) (*entity.Designation, error) {
	d, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find designation: %w", err)
	}
	if d == nil {
		return nil, errors.NewNotFound("designation not found")
	}
	d.Rename(name)
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to update designation: %w", err)
	}
	return d, nil
}

func (uc *DesignationUsecase) Delete(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete designation: %w", err)
	}
	return nil
}
