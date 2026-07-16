package delivery

import "github.com/gofiber/fiber/v3"

func (h *LeaveHandler) RegisterRoutes(r fiber.Router, authMw, adminMw fiber.Handler) {
	l := r.Group("/leave/types")
	l.Get("/", h.ListTypes)
	l.Get("/:id", h.GetType)
	l.Post("/", authMw, adminMw, h.CreateType)
	l.Put("/:id", authMw, adminMw, h.UpdateType)
	l.Delete("/:id", authMw, adminMw, h.DeleteType)

	r.Get("/leave/type-options", authMw, h.ListTypeOptions)

	b := r.Group("/leave/balances", authMw)
	b.Get("/", h.GetBalance)

	s := r.Group("/leave/submissions", authMw)
	s.Post("/", h.Submit)
	s.Get("/", h.ListMySubmissions)
	s.Get("/:id/attachment", h.GetAttachment)
	s.Get("/:id", h.GetSubmission)
	s.Post("/:id/cancel", h.CancelSubmission)

	admin := r.Group("/leave/admin", authMw, adminMw)
	admin.Get("/submissions", h.ListAllSubmissions)
	admin.Post("/submissions/:id/approve", h.ApproveSubmission)
	admin.Post("/submissions/:id/reject", h.RejectSubmission)
	admin.Get("/balances", h.ListBalances)
	admin.Put("/balances", h.UpdateBalance)
}
