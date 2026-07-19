package delivery

import "github.com/gofiber/fiber/v3"

func (h *AttendanceHandler) RegisterRoutes(r fiber.Router, authMw, adminMw fiber.Handler) {
	p := r.Group("/attendance/punch", authMw)
	p.Post("/in", h.PunchIn)
	p.Post("/out", h.PunchOut)
	p.Get("/today", h.PunchToday)
	p.Get("/history", h.PunchHistory)

	r.Get("/attendance/list", authMw, adminMw, h.DailyList)
	r.Get("/attendance/recap", authMw, adminMw, h.Recap)
	r.Get("/attendance/mine", authMw, h.MyAttendance)
	r.Get("/attendance/mine/history", authMw, h.MyAttendanceHistory)
	r.Get("/attendance/employee/:id/history", authMw, h.EmployeeAttendanceHistory)
	r.Get("/attendance/mine/stats", authMw, h.MyAttendanceStats)
	r.Get("/attendance/:id", authMw, adminMw, h.GetDetail)

	r.Post("/attendance/corrections", authMw, adminMw, h.CorrectionCreate)
	r.Get("/attendance/corrections", authMw, adminMw, h.CorrectionList)
	r.Delete("/attendance/corrections/:id", authMw, adminMw, h.CorrectionDelete)

	r.Get("/attendance/events", authMw, adminMw, h.PunchEvents)
}
