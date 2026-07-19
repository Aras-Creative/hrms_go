package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/audit/entity"
	"hrms/internal/audit/models"
	"hrms/internal/audit/repository"
)

type AuditUsecase struct {
	repo repository.AuditRepository
}

func NewAuditUsecase(repo repository.AuditRepository) *AuditUsecase {
	return &AuditUsecase{repo: repo}
}

func (uc *AuditUsecase) Log(ctx context.Context, data models.AuditLogData) error {
	e := entity.NewAuditEntry(
		data.Action,
		data.ActorID,
		data.Resource,
		data.ResourceID,
		data.TargetID,
		data.Payload,
		data.IP,
		data.UserAgent,
	)

	if err := uc.repo.Create(ctx, e); err != nil {
		return fmt.Errorf("audit log: %w", err)
	}
	return nil
}
func (uc *AuditUsecase) ListByResource(ctx context.Context, resource, resourceID string) ([]*entity.AuditEntry, error) {
	items, _, err := uc.repo.List(ctx, models.AuditFilter{
		Resource:   resource,
		ResourceID: resourceID,
		PerPage:    100,
	})
	if err != nil {
		return nil, fmt.Errorf("audit list by resource: %w", err)
	}
	return items, nil
}

func (uc *AuditUsecase) ListByResourceWithActor(ctx context.Context, resource, resourceID string) ([]*models.AuditEntryWithActor, error) {
	return uc.repo.ListByResourceWithActor(ctx, resource, resourceID)
}

func (uc *AuditUsecase) List(ctx context.Context, filter models.AuditFilter) ([]*entity.AuditEntry, int64, error) {
	items, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("audit list: %w", err)
	}
	return items, total, nil
}

func (uc *AuditUsecase) ListByResourceActionsAndDateRange(ctx context.Context, resource, resourceID string, actions []string, from, to time.Time) ([]*models.AuditEntryWithActor, error) {
	return uc.repo.ListByResourceActionsAndDateRange(ctx, resource, resourceID, actions, from, to)
}
