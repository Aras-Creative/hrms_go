package delivery

import "github.com/gofiber/fiber/v3"

func (h *ContractHandler) RegisterRoutes(r fiber.Router, authMw, adminMw fiber.Handler) {
	ct := r.Group("/contracts/templates")
	ct.Get("/", authMw, h.ListTemplates)
	ct.Get("/:id", authMw, h.GetTemplate)
	ct.Get("/:id/prefill", authMw, h.GetTemplatePrefill)
	ct.Post("/", authMw, adminMw, h.CreateTemplate)
	ct.Put("/:id", authMw, adminMw, h.UpdateTemplate)
	ct.Delete("/:id", authMw, adminMw, h.DeleteTemplate)

	r.Get("/contracts", authMw, adminMw, h.ListContracts)
	r.Get("/contracts/me", authMw, h.MyContracts)
	r.Get("/contracts/me/active", authMw, h.MyActiveContract)
	r.Get("/contracts/count-soon-expired", authMw, adminMw, h.CountSoonExpired)
	r.Get("/contracts/me/pending", authMw, h.PendingContracts)
	r.Get("/contracts/employee/:id", authMw, adminMw, h.GetEmployeeContract)
	r.Get("/contracts/:id", authMw, h.GetContract)
	r.Get("/contracts/:id/draft", authMw, adminMw, h.GetDraftContract)
	r.Put("/contracts/:id", authMw, adminMw, h.UpdateDraftContract)
	r.Get("/contracts/:id/preview", authMw, h.PreviewContract)
	r.Get("/contracts/:id/download", authMw, h.DownloadContract)
	r.Post("/contracts/:id/generate-pdf", authMw, adminMw, h.GeneratePDF)
	r.Post("/contracts/:id/terminate", authMw, adminMw, h.TerminateContract)
	r.Post("/contracts", authMw, adminMw, h.CreateContract)
	r.Post("/contracts/active", authMw, adminMw, h.CheckActiveContracts)
	r.Post("/contracts/sign", authMw, adminMw, h.SignByFirstParty)
	r.Post("/contracts/sign/second-party", authMw, h.SignBySecondParty)
	r.Delete("/contracts/:id", authMw, adminMw, h.DeleteContract)
}
