package adapter

import (
	"context"

	notifUc "hrms/internal/notification/usecase"
)

type NotifierAdapter struct {
	uc *notifUc.NotificationUsecase
}

func NewNotifierAdapter(uc *notifUc.NotificationUsecase) *NotifierAdapter {
	return &NotifierAdapter{uc: uc}
}

func (a *NotifierAdapter) Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error {
	return a.uc.Notify(ctx, userID, ntype, title, body, resource, resourceID)
}
