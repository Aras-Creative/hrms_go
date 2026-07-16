package delivery

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"hrms/internal/dashboard/usecase"
	errors "hrms/internal/pkg/apperror"
	response "hrms/internal/pkg/api"
)

type DashboardHandler struct {
	uc *usecase.DashboardUsecase
}

func NewDashboardHandler(uc *usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

func (h *DashboardHandler) GetMetrics(c fiber.Ctx) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	from := today
	if s := c.Query("from"); s != "" {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid from date, expected YYYY-MM-DD"))
		}
		from = d
	}

	to := today
	if s := c.Query("to"); s != "" {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid to date, expected YYYY-MM-DD"))
		}
		to = d
	}

	if from.After(to) {
		return response.Error(c, errors.NewInvalidInput("from cannot be after to"))
	}

	m, err := h.uc.GetMetrics(c.RequestCtx(), from, to)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toMetricsResponse(m))
}

func (h *DashboardHandler) GetAttendanceTrends(c fiber.Ctx) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	from := today
	if s := c.Query("from"); s != "" {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid from date, expected YYYY-MM-DD"))
		}
		from = d
	}

	to := today
	if s := c.Query("to"); s != "" {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid to date, expected YYYY-MM-DD"))
		}
		to = d
	}

	if from.After(to) {
		return response.Error(c, errors.NewInvalidInput("from cannot be after to"))
	}

	trends, err := h.uc.GetAttendanceTrends(c.RequestCtx(), from, to)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toTrendsResponse(trends))
}
