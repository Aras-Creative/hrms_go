package adapter

import (
	"context"

	"hrms/internal/audit/models"
	"hrms/internal/audit/usecase"
)

const (
	ActionCompCreate      = "payroll.compensation.create"
	ActionBenefitCreate   = "payroll.benefit.create"
	ActionDeductionCreate = "payroll.deduction.create"
	ActionPeriodCreate    = "payroll.period.create"
	ActionPeriodProcess   = "payroll.period.process"
	ActionPeriodClose     = "payroll.period.close"
	ActionSetup           = "payroll.setup"
	ActionPayslipCreate   = "payroll.payslip.create"
)

type AuditLogger struct {
	uc *usecase.AuditUsecase
}

func NewAuditLogger(uc *usecase.AuditUsecase) *AuditLogger {
	return &AuditLogger{uc: uc}
}

// Compile-time check: AuditLogger satisfies the interface expected by the delivery layer.
// (No formal interface defined; used as concrete type. Check added for future refactoring.)

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
