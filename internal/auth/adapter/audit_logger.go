package adapter

import (
	"context"

	"hrms/internal/audit/models"
	"hrms/internal/audit/usecase"
)

const (
	ActionRegister        = "auth.register"
	ActionDeviceRevoke    = "auth.device.revoke"
	ActionLogout          = "auth.logout"
	ActionPasswordChange  = "auth.password.change"
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
	var targetIDPtr *string
	if targetID != "" {
		targetIDPtr = &targetID
	}
	return a.uc.Log(ctx, models.AuditLogData{
		ActorID:    actorID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		TargetID:   targetIDPtr,
		Payload:    payload,
		IP:         ip,
		UserAgent:  userAgent,
	})
}
