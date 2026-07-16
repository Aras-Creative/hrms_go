package usecase

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"hrms/internal/leave/repository"
)

type EmployeeFetcher interface {
	GetAllActiveIDs(ctx context.Context) ([]string, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
	FindByUserID(ctx context.Context, userID string) (string, error)
	FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error)
}

type UserNameResolver interface {
	FindNameByID(ctx context.Context, userID string) (string, error)
	FindAdminIDs(ctx context.Context) ([]string, error)
}

type AttendanceReprocessor interface {
	ReprocessDay(ctx context.Context, employeeID string, date time.Time) (skippedDueToCorrection bool, err error)
}

// HalfDayPunchHandler handles auto punch-in/out for half-day leave approvals.
type HalfDayPunchHandler interface {
	EnsureHalfDayPunches(ctx context.Context, employeeID string, date time.Time) (createdIn, createdOut bool, err error)
}

type LeaveUsecase struct {
	db                  *sqlx.DB
	leaveTypeRepo       repository.LeaveTypeRepository
	leaveBalanceRepo    repository.LeaveBalanceRepository
	submissionRepo      repository.LeaveSubmissionRepository
	employeeFetcher     EmployeeFetcher
	userResolver        UserNameResolver
	attendanceProcessor AttendanceReprocessor
	halfDayPunchHandler HalfDayPunchHandler
	log                 *logrus.Logger
}

func (uc *LeaveUsecase) ResolveActorName(ctx context.Context, userID string) (string, error) {
	return uc.userResolver.FindNameByID(ctx, userID)
}

func (uc *LeaveUsecase) FindAdminIDs(ctx context.Context) ([]string, error) {
	return uc.userResolver.FindAdminIDs(ctx)
}

func (uc *LeaveUsecase) FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error) {
	return uc.employeeFetcher.FindUserIDByEmployeeID(ctx, employeeID)
}

func NewLeaveUsecase(
	db *sqlx.DB,
	leaveTypeRepo repository.LeaveTypeRepository,
	leaveBalanceRepo repository.LeaveBalanceRepository,
	submissionRepo repository.LeaveSubmissionRepository,
	employeeFetcher EmployeeFetcher,
	userResolver UserNameResolver,
	attendanceProcessor AttendanceReprocessor,
	halfDayPunchHandler HalfDayPunchHandler,
	log *logrus.Logger,
) *LeaveUsecase {
	return &LeaveUsecase{
		db:                  db,
		leaveTypeRepo:       leaveTypeRepo,
		leaveBalanceRepo:    leaveBalanceRepo,
		submissionRepo:      submissionRepo,
		employeeFetcher:     employeeFetcher,
		userResolver:        userResolver,
		attendanceProcessor: attendanceProcessor,
		halfDayPunchHandler: halfDayPunchHandler,
		log:                 log,
	}
}
