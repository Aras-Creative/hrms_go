package repository

import (
	"context"

	"hrms/internal/employee/entity"
	"hrms/internal/employee/models"
	"hrms/internal/employee/numbergen"
)

type EmployeeRepository interface {
	Create(ctx context.Context, e *entity.Employee) error
	FindByID(ctx context.Context, id string) (*entity.Employee, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Employee, error)
	FindByIDWithDetails(ctx context.Context, id string) (*models.EmployeeResult, error)
	FindByUserID(ctx context.Context, userID string) (*entity.Employee, error)
	FindByUserIDWithDetails(ctx context.Context, userID string) (*models.MeResult, error)
	FindAllActiveIDs(ctx context.Context) ([]string, error)
	FindAllWithDetails(ctx context.Context, filter models.ListEmployeeInput) ([]*models.EmployeeListItem, int64, error)
	Update(ctx context.Context, e *entity.Employee) error
	Delete(ctx context.Context, id string) error

	numbergen.SequenceRepository
}
