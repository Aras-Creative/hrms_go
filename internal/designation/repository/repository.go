package repository

import (
	"context"

	"hrms/internal/designation/entity"
	"hrms/internal/designation/models"
)

type DesignationRepository interface {
	Create(ctx context.Context, d *entity.Designation) error
	FindByID(ctx context.Context, id string) (*entity.Designation, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Designation, error)
	FindByCode(ctx context.Context, code string) (*entity.Designation, error)
	FindAll(ctx context.Context) ([]models.DesignationReadModel, error)
	Update(ctx context.Context, d *entity.Designation) error
	Delete(ctx context.Context, id string) error
}
