package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
)

type LeaveTypeRepository interface {
	Create(ctx context.Context, lt *entity.LeaveType) error
	FindByID(ctx context.Context, id string) (*entity.LeaveType, error)
	FindAllActive(ctx context.Context) ([]*entity.LeaveType, error)
	Update(ctx context.Context, lt *entity.LeaveType) error
}

type LeaveBalanceRepository interface {
	WithTx(tx *sqlx.Tx) LeaveBalanceRepository
	Create(ctx context.Context, lb *entity.LeaveBalance) error
	FindByEmployeeAndTypeYear(ctx context.Context, employeeID, leaveTypeID string, year int) (*entity.LeaveBalance, error)
	Update(ctx context.Context, lb *entity.LeaveBalance) error
	FindAll(ctx context.Context, filter BalanceFilter) ([]*models.LeaveBalance, int64, error)
	ConsumeBalance(ctx context.Context, id string, days int) error
}

type LeaveSubmissionRepository interface {
	WithTx(tx *sqlx.Tx) LeaveSubmissionRepository
	Create(ctx context.Context, s *entity.LeaveSubmission) error
	FindByID(ctx context.Context, id string) (*entity.LeaveSubmission, error)
	FindByEmployeeID(ctx context.Context, filter LeaveSubmissionFilter) ([]*entity.LeaveSubmission, int64, error)
	FindAll(ctx context.Context, filter SubmissionFilter) ([]*models.LeaveSubmission, int64, error)
	HasOverlap(ctx context.Context, employeeID string, startDate, endDate time.Time, excludeID string) (bool, error)
	HasApprovedLeave(ctx context.Context, employeeID string, date time.Time) (bool, *string, error)
	Update(ctx context.Context, s *entity.LeaveSubmission) error
}

type BalanceFilter struct {
	LeaveTypeID string
	Search      string
	Year        int
	Page        int
	PerPage     int
}

type LeaveSubmissionFilter struct {
	EmployeeID string
	Status     string
	Page       int
	PerPage    int
}

type SubmissionFilter struct {
	Status    string
	Search    string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PerPage   int
}

// LeaveSubmissionRepository is defined above with WithTx.
