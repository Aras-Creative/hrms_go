package repository

import (
	"context"

	"hrms/internal/notification/entity"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) error
	List(ctx context.Context, userID string, page, perPage int) ([]*entity.Notification, int64, error)
	UnreadCount(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	DeleteOlderThan(ctx context.Context, ttl string) error
}
