package delivery

import "github.com/gofiber/fiber/v3"

func (h *ScheduleHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
	s := r.Group("/schedule")

	// Employee-facing
	s.Get("/me/today", authMw, h.MyToday)

	// Admin patterns
	p := s.Group("/patterns", authMw)
	p.Get("/", h.ListPatterns)
	p.Get("/options", h.ListPatternOptions)
	p.Get("/:id", h.GetPattern)
	p.Post("/", h.CreatePattern)
	p.Put("/:id", h.UpdatePattern)
	p.Delete("/:id", h.DeletePattern)

	// Assignments
	a := s.Group("/assign", authMw)
	a.Post("/", h.Assign)
	a.Get("/:employeeId", h.GetActivePattern)
	a.Get("/:employeeId/history", h.GetPatternHistory)

	// Overview
	s.Get("/overview", authMw, h.Overview)

	// Overrides
	o := s.Group("/overrides", authMw)
	o.Post("/", h.SetOverride)
	o.Get("/", h.ListOverrides)
	o.Get("/:id", h.GetOverride)
	o.Delete("/:id", h.DeleteOverride)
}
