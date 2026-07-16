package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms/internal/schedule/entity"
	"hrms/internal/schedule/repository"
	apperrors "hrms/internal/pkg/apperror"
)

type SetOverrideInputRange struct {
	EmployeeID   string
	DateFrom     time.Time
	DateTo       time.Time
	IsWorkingDay bool
	StartTime    *string
	EndTime      *string
	Reason       *string
}

type ScheduleOverrideUsecase struct {
	repo repository.ScheduleOverrideRepository
}

func NewScheduleOverrideUsecase(repo repository.ScheduleOverrideRepository) *ScheduleOverrideUsecase {
	return &ScheduleOverrideUsecase{repo: repo}
}

func (uc *ScheduleOverrideUsecase) SetRange(ctx context.Context, input SetOverrideInputRange) ([]*entity.EmployeeScheduleOverride, error) {
	if input.DateTo.Before(input.DateFrom) {
		return nil, apperrors.NewInvalidInput("date_to must be after date_from")
	}

	var results []*entity.EmployeeScheduleOverride
	for d := input.DateFrom; !d.After(input.DateTo); d = d.AddDate(0, 0, 1) {
		o := entity.NewEmployeeScheduleOverride(
			input.EmployeeID, d, input.IsWorkingDay,
			input.StartTime, input.EndTime, input.Reason,
		)
		if err := uc.repo.Upsert(ctx, o); err != nil {
			return nil, fmt.Errorf("failed to set override for %s: %w", d.Format("2006-01-02"), err)
		}
		results = append(results, o)
	}
	return results, nil
}

func (uc *ScheduleOverrideUsecase) GetByID(ctx context.Context, id string) (*entity.EmployeeScheduleOverride, error) {
	o, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find override: %w", err)
	}
	if o == nil {
		return nil, apperrors.NewNotFound("schedule override not found")
	}
	return o, nil
}

func (uc *ScheduleOverrideUsecase) ListByEmployee(ctx context.Context, employeeID string, from, to string) ([]*entity.EmployeeScheduleOverride, error) {
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return nil, apperrors.NewInvalidInput(err.Error())
	}
	return uc.repo.FindByEmployeeAndDateRange(ctx, employeeID, fromTime, toTime)
}

func (uc *ScheduleOverrideUsecase) ListAll(ctx context.Context, from, to string) ([]*entity.EmployeeScheduleOverride, error) {
	fromTime, toTime, err := parseDateRange(from, to)
	if err != nil {
		return nil, apperrors.NewInvalidInput(err.Error())
	}
	return uc.repo.FindByDateRange(ctx, fromTime, toTime)
}

func (uc *ScheduleOverrideUsecase) Remove(ctx context.Context, id string) error {
	err := uc.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrOverrideNotFound) {
			return apperrors.NewNotFound("schedule override not found")
		}
		return fmt.Errorf("failed to delete override: %w", err)
	}
	return nil
}

func parseDateRange(from, to string) (time.Time, time.Time, error) {
	var fromTime, toTime time.Time
	var err error

	if from != "" {
		fromTime, err = tryParseDate(from)
		if err != nil {
			return fromTime, toTime, fmt.Errorf("invalid from date: %w", err)
		}
	} else {
		fromTime = time.Now().AddDate(0, -1, 0)
	}

	if to != "" {
		toTime, err = tryParseDate(to)
		if err != nil {
			return fromTime, toTime, fmt.Errorf("invalid to date: %w", err)
		}
	} else {
		toTime = time.Now().AddDate(0, 1, 0)
	}

	return fromTime, toTime, nil
}
