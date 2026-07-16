package delivery

import "github.com/gofiber/fiber/v3"

func (h *DashboardHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
	r.Get("/dashboard/metrics", authMw, h.GetMetrics)
	r.Get("/dashboard/attendance-trends", authMw, h.GetAttendanceTrends)
}
