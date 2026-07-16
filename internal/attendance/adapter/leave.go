package adapter

import (
	"context"
	"time"

	attendanceUc "hrms/internal/attendance/usecase"
	leaveRepo "hrms/internal/leave/repository"
)

type LeaveFetcherAdapter struct {
	repo leaveRepo.LeaveSubmissionRepository
}

func NewLeaveFetcherAdapter(repo leaveRepo.LeaveSubmissionRepository) *LeaveFetcherAdapter {
	return &LeaveFetcherAdapter{repo: repo}
}

func (a *LeaveFetcherAdapter) HasApprovedLeave(ctx context.Context, employeeID string, date time.Time) (bool, *string, error) {
	return a.repo.HasApprovedLeave(ctx, employeeID, date)
}

var _ attendanceUc.LeaveFetcher = (*LeaveFetcherAdapter)(nil)
