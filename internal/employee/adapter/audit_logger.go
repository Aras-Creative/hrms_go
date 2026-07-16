package adapter

import (
	"context"

	"hrms/internal/audit/models"
	"hrms/internal/audit/usecase"
)

const (
	ActionCreate       = "employee.create"
	ActionUpdate       = "employee.update"
	ActionPhotoUpdate  = "employee.profile-photo.update"
)

type AuditLogger struct {
	uc *usecase.AuditUsecase
}

func NewAuditLogger(uc *usecase.AuditUsecase) *AuditLogger {
	return &AuditLogger{uc: uc}
}

func (a *AuditLogger) Log(
	ctx context.Context,
	actorID, resourceID, targetID, action, ip, userAgent string,
	payload map[string]any,
) error {
	return a.uc.Log(ctx, models.AuditLogData{
		ActorID:    actorID,
		Action:     action,
		Resource:   "employee",
		ResourceID: resourceID,
		TargetID:   &targetID,
		Payload:    payload,
		IP:         ip,
		UserAgent:  userAgent,
	})
}
