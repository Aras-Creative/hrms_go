package delivery

import (
	"strconv"

	"hrms/internal/audit/models"
	"hrms/internal/audit/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

type AuditHandler struct {
	uc *usecase.AuditUsecase
}

func NewAuditHandler(uc *usecase.AuditUsecase) *AuditHandler {
	return &AuditHandler{uc: uc}
}

func (h *AuditHandler) MyActivityLogs(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	entries, total, err := h.uc.List(c.RequestCtx(), models.AuditFilter{
		ActorID: userID,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return response.Error(c, err)
	}

	items := make([]ActivityLogResponse, 0, len(entries))
	for _, e := range entries {
		items = append(items, toActivityLogResponse(e))
	}
	return response.Paginate(c, items, page, perPage, total)
}
