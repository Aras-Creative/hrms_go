package usecase

import (
	"context"
	"time"
)

type EmployeeFetcher interface {
	FindByUserID(ctx context.Context, userID string) (employeeID, employeeName string, err error)
	FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error)
}

type EmployeeExistenceChecker interface {
	ExistsByID(ctx context.Context, id string) (bool, error)
}

type LeaveFetcher interface {
	HasApprovedLeave(ctx context.Context, employeeID string, date time.Time) (bool, *string, error)
}

type ResolvedSchedule struct {
	ExpectedStartTime  *string
	ExpectedEndTime    *string
	Source             string
	ScheduleOverrideID *string
	OverrideIsWorking  *bool
	WorkingType        string
}

type ScheduleResolver interface {
	ResolveRange(ctx context.Context, employeeID string, from, to time.Time) (map[string]map[string]*ResolvedSchedule, error)
}

type CorrectionAuditInfo struct {
	ActorID    string
	ActorName  string
	Action     string
	Payload    map[string]any
	CreatedAt  time.Time
}

type CorrectionAuditFetcher interface {
	FetchCorrectionLogs(ctx context.Context, employeeID string, from, to time.Time) (map[string]*CorrectionAuditInfo, error)
}
