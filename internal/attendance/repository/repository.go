package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/attendance/entity"
)

// txContext is satisfied by both *sqlx.DB and *sqlx.Tx.
type txContext interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type DailyAttendanceRepository interface {
	WithTx(tx *sqlx.Tx) DailyAttendanceRepository
	Upsert(ctx context.Context, da *entity.DailyAttendance) error
	FindByID(ctx context.Context, id string) (*entity.DailyAttendance, error)
	FindByEmployeeAndDateRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.DailyAttendance, error)
	FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*entity.DailyAttendance, error)
	ComputeDaily(ctx context.Context, employeeID string, date time.Time) (*DailyComputationRow, error)
	ComputeRange(ctx context.Context, from, to time.Time) ([]DailyComputationRow, error)
	FindAllPaginated(ctx context.Context, searchName, status, designationID, isLate, isEarlyLeave string, from, to time.Time, page, perPage int) ([]*AdminAttendanceRow, int64, error)
	Recap(ctx context.Context, from, to time.Time, designationID string) ([]*RecapRow, error)
	FindActiveLeaveTypes(ctx context.Context) ([]LeaveTypeRow, error)
}

type PunchRepository interface {
	Create(ctx context.Context, p *entity.Punch) error
	FindByEmployeeAndDateRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.Punch, error)
	FindTodayByEmployee(ctx context.Context, employeeID string) ([]*entity.Punch, error)
}

type CorrectionRepository interface {
	WithTx(tx *sqlx.Tx) CorrectionRepository
	Create(ctx context.Context, c *entity.AttendanceCorrection) error
	Update(ctx context.Context, c *entity.AttendanceCorrection) error
	FindByID(ctx context.Context, id string) (*entity.AttendanceCorrection, error)
	FindByEmployeeAndDate(ctx context.Context, employeeID string, date time.Time) (*entity.AttendanceCorrection, error)
	FindAllPaginated(ctx context.Context, searchName string, from, to time.Time, page, perPage int) ([]*CorrectionViewRow, int64, error)
	Delete(ctx context.Context, id string) error
}

type CorrectionViewRow struct {
	CorrectionModel
	EmployeeName string `db:"employee_name"`
}
