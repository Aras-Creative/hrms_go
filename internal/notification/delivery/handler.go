package delivery

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	notifUc "hrms/internal/notification/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
)

type NotificationHandler struct {
	uc *notifUc.NotificationUsecase
}

func NewNotificationHandler(uc *notifUc.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

func (h *NotificationHandler) List(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(c.Query("per_page", "20"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 20
	}

	items, total, err := h.uc.List(c.RequestCtx(), userID, page, perPage)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, fiber.Map{
		"items": items,
		"total": total,
	})
}

func (h *NotificationHandler) UnreadCount(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	count, err := h.uc.UnreadCount(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, fiber.Map{"unread": count})
}

func (h *NotificationHandler) MarkRead(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}
	id, err := response.ParseParamID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}

	if err := h.uc.MarkRead(c.RequestCtx(), id, userID); err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, nil)
}

func (h *NotificationHandler) MarkAllRead(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	if err := h.uc.MarkAllRead(c.RequestCtx(), userID); err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, nil)
}
