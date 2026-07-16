package adapter

import (
	"context"

	"hrms/internal/audit/models"
	"hrms/internal/audit/usecase"
)

const (
	ActionCreate = "designation.create"
	ActionUpdate = "designation.update"
	ActionDelete = "designation.delete"
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
		Resource:   "designation",
		ResourceID: resourceID,
		TargetID:   &targetID,
		Payload:    payload,
		IP:         ip,
		UserAgent:  userAgent,
	})
}
