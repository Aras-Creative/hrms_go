package adapter

import (
	"context"

	"hrms/internal/audit/entity"
	"hrms/internal/audit/models"
	"hrms/internal/audit/usecase"
)

const (
	ActionSubmit        = "leave.submit"
	ActionApprove       = "leave.approve"
	ActionReject        = "leave.reject"
	ActionCancel        = "leave.cancel"
	ActionTypeCreate    = "leave.type.create"
	ActionTypeUpdate    = "leave.type.update"
	ActionTypeDelete    = "leave.type.delete"
	ActionBalanceUpdate = "leave.balance.update"
)

type AuditLogger struct {
	uc *usecase.AuditUsecase
}

func NewAuditLogger(uc *usecase.AuditUsecase) *AuditLogger {
	return &AuditLogger{uc: uc}
}

func (a *AuditLogger) Log(
	ctx context.Context,
	actorID, resource, resourceID, targetID, action, ip, userAgent string,
	payload map[string]any,
) error {
	return a.uc.Log(ctx, models.AuditLogData{
		ActorID:    actorID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		TargetID:   &targetID,
		Payload:    payload,
		IP:         ip,
		UserAgent:  userAgent,
	})
}

func (a *AuditLogger) ListByResource(ctx context.Context, resource, resourceID string) ([]*entity.AuditEntry, error) {
	return a.uc.ListByResource(ctx, resource, resourceID)
}
