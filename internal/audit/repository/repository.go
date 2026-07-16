package repository

import (
	"context"

	"hrms/internal/audit/entity"
	"hrms/internal/audit/models"
)

type AuditRepository interface {
	Create(ctx context.Context, e *entity.AuditEntry) error
	List(ctx context.Context, filter models.AuditFilter) ([]*entity.AuditEntry, int64, error)
	ListByResourceWithActor(ctx context.Context, resource, resourceID string) ([]*models.AuditEntryWithActor, error)
}
