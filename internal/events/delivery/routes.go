package delivery

import "github.com/gofiber/fiber/v3"

func (h *EventHandler) RegisterRoutes(r fiber.Router, sseAuthMw fiber.Handler) {
	r.Get("/events", sseAuthMw, h.Stream)
}
