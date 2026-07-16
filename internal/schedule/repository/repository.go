package repository

import (
	"context"
	"time"

	"hrms/internal/schedule/entity"
)

type WorkPatternRepository interface {
	Create(ctx context.Context, wp *entity.WorkingPattern) error
	FindByID(ctx context.Context, id string) (*entity.WorkingPattern, error)
	FindAll(ctx context.Context) ([]*entity.WorkingPattern, error)
	FindAllActive(ctx context.Context) ([]*entity.WorkingPattern, error)
	Update(ctx context.Context, wp *entity.WorkingPattern) error
}

type EmployeeWorkPatternRepository interface {
	Create(ctx context.Context, ewp *entity.EmployeeWorkPattern) error
	Upsert(ctx context.Context, ewp *entity.EmployeeWorkPattern) error
	FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*entity.EmployeeWorkPattern, error)
	FindActiveByEmployee(ctx context.Context, employeeID string) (*entity.EmployeeWorkPattern, error)
	FindHistoryByEmployee(ctx context.Context, employeeID string) ([]*entity.EmployeeWorkPattern, error)
	DeactivateCurrent(ctx context.Context, employeeID string, validTo time.Time) error
}

type ScheduleOverrideRepository interface {
	Upsert(ctx context.Context, o *entity.EmployeeScheduleOverride) error
	FindByID(ctx context.Context, id string) (*entity.EmployeeScheduleOverride, error)
	FindByEmployeeAndDateRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.EmployeeScheduleOverride, error)
	FindByDateRange(ctx context.Context, from, to time.Time) ([]*entity.EmployeeScheduleOverride, error)
	Delete(ctx context.Context, id string) error
	DeleteByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) error
	DeleteFutureOverridesByEmployee(ctx context.Context, employeeID string, since time.Time) error
	QueryOverview(ctx context.Context, p ScheduleOverviewParams) ([]ScheduleOverviewRow, error)
}
