package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"hrms/internal/attendance/entity"
	"hrms/internal/attendance/models"
	"hrms/internal/attendance/repository"
	errors "hrms/internal/pkg/apperror"
	"hrms/internal/pkg/sse"
	"hrms/internal/pkg/timeutil"
)

type PunchUsecase struct {
	repo         repository.PunchRepository
	processor    *DailyProcessor
	leaveFetcher LeaveFetcher
	hub          *sse.Hub
}

func NewPunchUsecase(repo repository.PunchRepository, processor *DailyProcessor, leaveFetcher LeaveFetcher, hub *sse.Hub) *PunchUsecase {
	return &PunchUsecase{repo: repo, processor: processor, leaveFetcher: leaveFetcher, hub: hub}
}

func (uc *PunchUsecase) PunchIn(ctx context.Context, input models.PunchInput) (*entity.Punch, error) {
	return uc.punch(ctx, entity.PunchIn, input)
}

func (uc *PunchUsecase) PunchOut(ctx context.Context, input models.PunchInput) (*entity.Punch, error) {
	return uc.punch(ctx, entity.PunchOut, input)
}

func (uc *PunchUsecase) punch(ctx context.Context, punchType entity.PunchType, input models.PunchInput) (*entity.Punch, error) {
	now := time.Now()
	today := entity.LocalDate(now)

	onLeave, _, err := uc.leaveFetcher.HasApprovedLeave(ctx, input.EmployeeID, today)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to check leave status: %v", err))
	}
	if onLeave {
		return nil, errors.NewForbidden("cannot punch while on approved leave")
	}

	p := entity.NewPunch(input.EmployeeID, punchType, now)
	if err := uc.repo.Create(ctx, p); err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to create punch: %v", err))
	}

	da, err := uc.processor.ProcessDaily(ctx, input.EmployeeID, p.Date)
	if err != nil {
		slog.Error("failed to process daily attendance after punch",
			"employee_id", input.EmployeeID, "error", err)
	}

	if uc.hub != nil && da != nil {
		evt := models.PunchEvent{
			EmployeeID:       input.EmployeeID,
			PunchType:        string(punchType),
			Timestamp:        p.Timestamp,
			Status:           string(da.Status),
			FirstPunchIn:     da.FirstPunchIn,
			LastPunchOut:     da.LastPunchOut,
			LateMinutes:      da.LateMinutes(),
			TotalWorkSeconds: da.TotalWorkSeconds,
		}
		if data, err := json.Marshal(evt); err == nil {
			uc.hub.Publish("punches", string(data))
		}
	}

	return p, nil
}

func (uc *PunchUsecase) GetToday(ctx context.Context, employeeID string) ([]*entity.Punch, error) {
	return uc.repo.FindTodayByEmployee(ctx, employeeID)
}

func (uc *PunchUsecase) GetHistory(ctx context.Context, input models.PunchHistoryInput) ([]*entity.Punch, error) {
	from, to, err := timeutil.ParseDateRange(input.From, input.To)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	return uc.repo.FindByEmployeeAndDateRange(ctx, input.EmployeeID, from, to)
}
