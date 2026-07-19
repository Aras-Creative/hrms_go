package usecase

import (
	"context"
	"math"
	"time"

	"hrms/internal/dashboard/entity"
	"hrms/internal/dashboard/repository"
)

type DashboardUsecase struct {
	repo repository.DashboardRepository
}

func NewDashboardUsecase(repo repository.DashboardRepository) *DashboardUsecase {
	return &DashboardUsecase{repo: repo}
}

func (uc *DashboardUsecase) GetMetrics(ctx context.Context, from, to time.Time) (*entity.DashboardMetrics, error) {
	mc, err := uc.repo.CountMetrics(ctx, from, to)
	if err != nil {
		return nil, err
	}

	days := int(to.Sub(from).Hours()/24) + 1
	capacity := mc.TotalEmployees * days

	var presentPct, latePct float64
	if capacity > 0 {
		presentPct = math.Round(float64(mc.Present)*1000/float64(capacity)) / 10
		latePct = math.Round(float64(mc.Late)*1000/float64(capacity)) / 10
	}

	return &entity.DashboardMetrics{
		TotalEmployees:  mc.TotalEmployees,
		ActiveContracts: mc.ActiveContracts,
		Present:         mc.Present,
		Absent:          mc.Absent,
		PresentPct:      presentPct,
		Late:            mc.Late,
		LatePct:         latePct,
		OnLeave:         mc.OnLeave,
		From:            from,
		To:              to,
	}, nil
}

func (uc *DashboardUsecase) GetAttendanceTrends(ctx context.Context, from, to time.Time) ([]entity.AttendanceTrend, error) {
	rows, err := uc.repo.GetTrends(ctx, from, to)
	if err != nil {
		return nil, err
	}

	result := make([]entity.AttendanceTrend, 0, len(rows))
	for _, r := range rows {
		result = append(result, entity.AttendanceTrend{
			Date: r.Date, Present: r.Present,
			OnLeave: r.OnLeave, Absent: r.Absent,
		})
	}
	return result, nil
}
