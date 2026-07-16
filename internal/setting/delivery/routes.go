package delivery

import "github.com/gofiber/fiber/v3"

func (h *SettingHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler, adminMw fiber.Handler) {
	r.Get("/settings", authMw, h.Get)
	r.Put("/settings", authMw, adminMw, h.Update)
}
