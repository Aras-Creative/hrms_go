package adapter

import (
	"context"
	"time"

	auditUC "hrms/internal/audit/usecase"
	"hrms/internal/attendance/usecase"
)

type CorrectionAuditFetcherAdapter struct {
	auditUC *auditUC.AuditUsecase
}

func NewCorrectionAuditFetcherAdapter(auditUC *auditUC.AuditUsecase) *CorrectionAuditFetcherAdapter {
	return &CorrectionAuditFetcherAdapter{auditUC: auditUC}
}

func (a *CorrectionAuditFetcherAdapter) FetchCorrectionLogs(ctx context.Context, employeeID string, from, to time.Time) (map[string]*usecase.CorrectionAuditInfo, error) {
	actions := []string{
		ActionCorrection,
		ActionCorrectionUpdate,
	}

	logs, err := a.auditUC.ListByResourceActionsAndDateRange(ctx, "employee", employeeID, actions, from, to)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*usecase.CorrectionAuditInfo, len(logs))
	for _, log := range logs {
		dateStr, _ := log.Payload["date"].(string)
		if dateStr == "" {
			continue
		}
		if _, exists := result[dateStr]; !exists {
			result[dateStr] = &usecase.CorrectionAuditInfo{
				ActorID:   log.ActorID,
				ActorName: log.ActorName,
				Action:    log.Action,
				Payload:   log.Payload,
				CreatedAt: log.CreatedAt,
			}
		}
	}

	return result, nil
}
