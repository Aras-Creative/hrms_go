package delivery

import "github.com/gofiber/fiber/v3"

func (h *DesignationHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
	d := r.Group("/designations")
	d.Get("/options", h.Options)
	d.Get("/", h.FindAll)
	d.Get("/:id", h.FindByID)
	d.Post("/", authMw, h.Create)
	d.Put("/:id", authMw, h.Update)
	d.Delete("/:id", authMw, h.Delete)
}
