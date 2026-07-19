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

	date := today
	if s := c.Query("date"); s != "" {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid date, expected YYYY-MM-DD"))
		}
		date = d
	}

	m, err := h.uc.GetMetrics(c.RequestCtx(), date, date)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toMetricsResponse(m))
}

func (h *DashboardHandler) GetAttendanceTrends(c fiber.Ctx) error {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		return response.Error(c, errors.NewInvalidInput("from and to are required, expected YYYY-MM-DD"))
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid from, expected YYYY-MM-DD"))
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid to, expected YYYY-MM-DD"))
	}

	trends, err := h.uc.GetAttendanceTrends(c.RequestCtx(), from, to)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toTrendsResponse(trends))
}
