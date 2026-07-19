package repository

import (
	"context"
	"time"

	"hrms/internal/audit/entity"
	"hrms/internal/audit/models"
)

type AuditRepository interface {
	Create(ctx context.Context, e *entity.AuditEntry) error
	List(ctx context.Context, filter models.AuditFilter) ([]*entity.AuditEntry, int64, error)
	ListByResourceWithActor(ctx context.Context, resource, resourceID string) ([]*models.AuditEntryWithActor, error)
	ListByResourceActionsAndDateRange(ctx context.Context, resource, resourceID string, actions []string, from, to time.Time) ([]*models.AuditEntryWithActor, error)
}
