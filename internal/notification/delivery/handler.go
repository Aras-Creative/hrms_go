package delivery

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	notifUc "hrms/internal/notification/usecase"
	response "hrms/internal/pkg/api"
)

type NotificationHandler struct {
	uc *notifUc.NotificationUsecase
}

func NewNotificationHandler(uc *notifUc.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

func (h *NotificationHandler) List(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

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
	userID := c.Locals("user_id").(string)

	count, err := h.uc.UnreadCount(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, fiber.Map{"unread": count})
}

func (h *NotificationHandler) MarkRead(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	id := c.Params("id")

	if err := h.uc.MarkRead(c.RequestCtx(), id, userID); err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, nil)
}

func (h *NotificationHandler) MarkAllRead(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.uc.MarkAllRead(c.RequestCtx(), userID); err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, nil)
}
