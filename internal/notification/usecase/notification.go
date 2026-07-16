package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"hrms/internal/notification/entity"
	"hrms/internal/notification/repository"
	"hrms/internal/pkg/sse"
)

type NotificationUsecase struct {
	repo repository.NotificationRepository
	hub  *sse.Hub
}

func NewNotificationUsecase(repo repository.NotificationRepository, hub *sse.Hub) *NotificationUsecase {
	return &NotificationUsecase{repo: repo, hub: hub}
}

type NotifyEvent struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
}

func (uc *NotificationUsecase) Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error {
	n := entity.NewNotification(userID, ntype, title, body, resource, resourceID)
	if err := uc.repo.Create(ctx, n); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	evt := NotifyEvent{
		ID:         n.ID,
		UserID:     n.UserID,
		Type:       n.Type,
		Title:      n.Title,
		Body:       n.Body,
		Resource:   n.Resource,
		ResourceID: n.ResourceID,
		IsRead:     n.IsRead,
		CreatedAt:  n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	data, _ := json.Marshal(evt)
	uc.hub.Publish("notifications:"+userID, string(data))

	return nil
}

func (uc *NotificationUsecase) List(ctx context.Context, userID string, page, perPage int) ([]*entity.Notification, int64, error) {
	return uc.repo.List(ctx, userID, page, perPage)
}

func (uc *NotificationUsecase) UnreadCount(ctx context.Context, userID string) (int64, error) {
	return uc.repo.UnreadCount(ctx, userID)
}

func (uc *NotificationUsecase) MarkRead(ctx context.Context, id, userID string) error {
	return uc.repo.MarkRead(ctx, id, userID)
}

func (uc *NotificationUsecase) MarkAllRead(ctx context.Context, userID string) error {
	return uc.repo.MarkAllRead(ctx, userID)
}
