package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/leave/entity"
)

// txContext is satisfied by both *sqlx.DB and *sqlx.Tx.
type txContext interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

var (
	queryLeaveTypeByID       = `SELECT id, name, default_days, is_paid, is_unlimited, is_half_day, is_active, created_at, updated_at FROM leave_types WHERE id = $1`
	queryLeaveTypesAllActive = `SELECT id, name, default_days, is_paid, is_unlimited, is_half_day, is_active, created_at, updated_at FROM leave_types WHERE is_active = true`
	queryUpdateLeaveType     = `UPDATE leave_types SET name = $1, default_days = $2, is_paid = $3, is_unlimited = $4, is_half_day = $5, is_active = $6, updated_at = $7 WHERE id = $8`
	querySelectLeaveType     = `SELECT id, name, default_days, is_paid, is_unlimited, is_half_day, is_active, created_at, updated_at FROM leave_types`
	querySelectLeaveBalance  = `SELECT id, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at FROM leave_balances`
	queryLeaveBalanceByEmpType = `SELECT id, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at FROM leave_balances WHERE employee_id = $1 AND leave_type_id = $2 AND year = $3`
)

func modelToLeaveType(m *LeaveTypeModel) *entity.LeaveType {
	return entity.ReconstituteLeaveType(m.ID, m.Name, m.DefaultDays, m.IsPaid, m.IsUnlimited, m.IsHalfDay, m.IsActive, m.CreatedAt, m.UpdatedAt)
}

func modelToLeaveBalance(m *LeaveBalanceModel) *entity.LeaveBalance {
	return entity.ReconstituteLeaveBalance(m.ID, m.EmployeeID, m.LeaveTypeID, m.Year, m.TotalDays, m.UsedDays, m.CreatedAt, m.UpdatedAt)
}

type LeaveTypeModel struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	DefaultDays int       `db:"default_days"`
	IsPaid      bool      `db:"is_paid"`
	IsUnlimited bool      `db:"is_unlimited"`
	IsHalfDay   bool      `db:"is_half_day"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type LeaveSubmissionModel struct {
	ID           string     `db:"id"`
	EmployeeID   string     `db:"employee_id"`
	LeaveTypeID  string     `db:"leave_type_id"`
	StartDate    time.Time  `db:"start_date"`
	EndDate      time.Time  `db:"end_date"`
	Days         int        `db:"days"`
	Reason       string     `db:"reason"`
	AttachmentID *string    `db:"attachment_id"`
	Status       string     `db:"status"`
	ApprovedBy   *string    `db:"approved_by"`
	ApprovedAt   *time.Time `db:"approved_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type LeaveSubmissionWithEmployeeModel struct {
	ID             string     `db:"id"`
	EmployeeID     string     `db:"employee_id"`
	ProfilePhotoID *string    `db:"profile_photo_id"`
	LeaveTypeID    string     `db:"leave_type_id"`
	LeaveTypeName  string     `db:"leave_type_name"`
	StartDate      time.Time  `db:"start_date"`
	EndDate        time.Time  `db:"end_date"`
	Days           int        `db:"days"`
	Reason         string     `db:"reason"`
	AttachmentID   *string    `db:"attachment_id"`
	Status         string     `db:"status"`
	ApprovedBy     *string    `db:"approved_by"`
	ApprovedAt     *time.Time `db:"approved_at"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	EmployeeName   string     `db:"employee_name"`
	EmployeeNumber string     `db:"employee_number"`
}

type LeaveBalanceModel struct {
	ID          string    `db:"id"`
	EmployeeID  string    `db:"employee_id"`
	LeaveTypeID string    `db:"leave_type_id"`
	Year        int       `db:"year"`
	TotalDays   int       `db:"total_days"`
	UsedDays    int       `db:"used_days"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type LeaveBalanceWithEmployeeModel struct {
	ID             string    `db:"id"`
	EmployeeID     string    `db:"employee_id"`
	LeaveTypeID    string    `db:"leave_type_id"`
	Year           int       `db:"year"`
	TotalDays      int       `db:"total_days"`
	UsedDays       int       `db:"used_days"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	EmployeeName   string    `db:"employee_name"`
	EmployeeNumber string    `db:"employee_number"`
	ProfilePhotoID *string   `db:"profile_photo_id"`
	LeaveTypeName  string    `db:"leave_type_name"`
}