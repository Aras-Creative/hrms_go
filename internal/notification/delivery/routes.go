package delivery

import "github.com/gofiber/fiber/v3"

func (h *NotificationHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
	n := r.Group("/notifications", authMw)
	n.Get("/", h.List)
	n.Get("/unread", h.UnreadCount)
	n.Put("/:id/read", h.MarkRead)
	n.Put("/read-all", h.MarkAllRead)
}
