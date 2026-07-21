package adapter

import (
	"context"

	"hrms/internal/audit/models"
	"hrms/internal/audit/usecase"
)

const (
	ActionCreate          = "contract.create"
	ActionUpdate          = "contract.update"
	ActionSignFirstParty  = "contract.sign-first-party"
	ActionSignSecondParty = "contract.sign-second-party"
	ActionTerminate       = "contract.terminate"
	ActionTemplateCreate  = "contract.template.create"
	ActionTemplateUpdate  = "contract.template.update"
	ActionTemplateDelete  = "contract.template.delete"
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
