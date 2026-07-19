package delivery

import "github.com/gofiber/fiber/v3"

func (h *EmployeeHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
	e := r.Group("/employees")
	e.Get("/", h.List)
	e.Get("/me", authMw, h.FindByUserID)
	e.Put("/me/profile-photo", authMw, h.UpdateMyProfilePhoto)
	e.Get("/peek-next-number", authMw, h.PeekNextNumber)
	e.Get("/:id", h.FindByID)
	e.Post("/", authMw, h.Create)
	e.Put("/designation", authMw, h.ChangeDesignation)
	e.Put("/:id", authMw, h.Upsert)
	e.Put("/:id/profile-photo", authMw, h.UpdateProfilePhoto)
	e.Put("/:id/contact", authMw, h.UpdateContact)
	e.Put("/:id/identity", authMw, h.UpdateIdentity)
	e.Put("/:id/bank", authMw, h.UpdateBank)
	e.Get("/me/completion", authMw, h.MyProfileCompletion)
}
