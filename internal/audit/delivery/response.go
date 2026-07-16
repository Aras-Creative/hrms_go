package delivery

import (
	"time"

	"hrms/internal/audit/entity"
)

type ActivityLogResponse struct {
	ID        string         `json:"id"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

func toActivityLogResponse(e *entity.AuditEntry) ActivityLogResponse {
	return ActivityLogResponse{
		ID:        e.ID,
		Action:    e.Action,
		Resource:  e.Resource,
		Payload:   e.Payload,
		CreatedAt: e.CreatedAt,
	}
}
