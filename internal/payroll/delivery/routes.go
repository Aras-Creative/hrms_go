package delivery

import "github.com/gofiber/fiber/v3"

func (h *PayrollHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler, adminMw fiber.Handler) {
	// Employee-facing routes (auth only, no admin required)
	my := r.Group("/payroll/my", authMw)
	my.Get("/pay-slips", h.GetMyPayslips)
	my.Get("/pay-slips/:period_id", h.GetMyPaySlip)
	my.Get("/pay-slips/:period_id/print", h.PrintMyPaySlip)

	// Admin routes
	p := r.Group("/payroll", authMw, adminMw)

	// Bulk setup
	p.Post("/setup", h.SetupEmployee)

	// Periods
	periods := p.Group("/periods")
	periods.Get("/", h.ListPeriods)
	periods.Post("/", h.CreatePeriod)
	periods.Post("/:id/process", h.ProcessPeriod)
	periods.Post("/:id/close", h.ClosePeriod)
	periods.Get("/:id/pay-slips", h.ListPaySlips)
	periods.Post("/:id/pay-slips", h.CreateManualPaySlip)
	periods.Get("/:id/overview", h.GetPeriodOverview)

	// Pay slips
	p.Get("/pay-slips/:id", h.GetPaySlip)
	p.Get("/pay-slips/:id/print", h.PrintPaySlip)

	// Compensation items (master)
	comp := p.Group("/compensation")
	comp.Get("/options", h.ListCompensationItemOptions)
	comp.Get("/", h.ListCompensationItems)
	comp.Post("/", h.CreateCompensationItem)

	// Benefit types (master)
	bnft := p.Group("/benefit")
	bnft.Get("/options", h.ListBenefitTypeOptions)
	bnft.Get("/", h.ListBenefitTypes)
	bnft.Post("/", h.CreateBenefitType)

	// Deduction types (master)
	ded := p.Group("/deduction")
	ded.Get("/options", h.ListDeductionTypeOptions)
	ded.Get("/", h.ListDeductionTypes)
	ded.Post("/", h.CreateDeductionType)

	// Employee payroll components
	p.Get("/employees/:id/components", h.GetEmployeeComponents)
}
