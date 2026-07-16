package delivery

import "github.com/gofiber/fiber/v3"

func (h *AuditHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
	r.Get("/mine/activity", authMw, h.MyActivityLogs)
}
