package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	response "hrms/internal/pkg/api"
	"hrms/internal/schedule/entity"
	"hrms/internal/schedule/repository"
	errors "hrms/internal/pkg/apperror"
)

type WorkPatternUsecase struct {
	repo repository.WorkPatternRepository
}

func NewWorkPatternUsecase(repo repository.WorkPatternRepository) *WorkPatternUsecase {
	return &WorkPatternUsecase{repo: repo}
}

func (uc *WorkPatternUsecase) Create(ctx context.Context, input CreateWorkingPatternInput) (*entity.WorkingPattern, error) {
	details := make([]entity.WorkingPatternDetail, 0, len(input.Details))
	for _, d := range input.Details {
		wt, _ := entity.ParseWorkingType(d.Type)
		details = append(details, entity.WorkingPatternDetail{
			DayOfWeek: entity.DayOfWeek(d.DayOfWeek),
			Type:      wt,
			StartTime: d.StartTime,
			EndTime:   d.EndTime,
		})
	}

	wp := entity.NewWorkingPattern(input.Name, input.Description, details)
	if err := uc.repo.Create(ctx, wp); err != nil {
		return nil, fmt.Errorf("failed to create work pattern: %w", err)
	}
	return wp, nil
}

func (uc *WorkPatternUsecase) GetByID(ctx context.Context, id string) (*entity.WorkingPattern, error) {
	wp, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find work pattern: %w", err)
	}
	if wp == nil {
		return nil, errors.NewNotFound("work pattern not found")
	}
	return wp, nil
}

func (uc *WorkPatternUsecase) GetAll(ctx context.Context) ([]*entity.WorkingPattern, error) {
	return uc.repo.FindAllActive(ctx)
}

func (uc *WorkPatternUsecase) Update(ctx context.Context, id string, input UpdateWorkingPatternInput) (*entity.WorkingPattern, error) {
	wp, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find work pattern: %w", err)
	}
	if wp == nil {
		return nil, errors.NewNotFound("work pattern not found")
	}

	if input.Name != "" {
		wp.Rename(input.Name)
	}
	wp.Description = input.Description

	wp.Details = make([]entity.WorkingPatternDetail, 0, len(input.Details))
	for _, d := range input.Details {
		wt, _ := entity.ParseWorkingType(d.Type)
		wp.Details = append(wp.Details, entity.WorkingPatternDetail{
			ID:               uuid.New().String(),
			WorkingPatternID: id,
			DayOfWeek:        entity.DayOfWeek(d.DayOfWeek),
			Type:             wt,
			StartTime:        d.StartTime,
			EndTime:          d.EndTime,
		})
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, fmt.Errorf("failed to update work pattern: %w", err)
	}
	return wp, nil
}

func (uc *WorkPatternUsecase) Delete(ctx context.Context, id string) error {
	wp, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find work pattern: %w", err)
	}
	if wp == nil {
		return errors.NewNotFound("work pattern not found")
	}
	if err := wp.Disable(); err != nil {
		return fmt.Errorf("disable work pattern: %w", err)
	}
	if err := uc.repo.Update(ctx, wp); err != nil {
		return fmt.Errorf("update work pattern: %w", err)
	}
	return nil
}

func (uc *WorkPatternUsecase) GetOptions(ctx context.Context) ([]response.Option, error) {
	list, err := uc.repo.FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list work patterns: %w", err)
	}
	opts := make([]response.Option, 0, len(list))
	for _, wp := range list {
		opts = append(opts, response.Option{
			Value: wp.ID,
			Label: wp.Name,
		})
	}
	return opts, nil
}
