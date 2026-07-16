package adapter

import (
	"context"

	leaveUc "hrms/internal/leave/usecase"
	emplUc "hrms/internal/employee/usecase"
)

type BalanceAssignerAdapter struct {
	uc *leaveUc.LeaveUsecase
}

func NewBalanceAssignerAdapter(uc *leaveUc.LeaveUsecase) *BalanceAssignerAdapter {
	return &BalanceAssignerAdapter{uc: uc}
}

func (a *BalanceAssignerAdapter) AssignBalances(ctx context.Context, employeeID string) error {
	return a.uc.AssignEmployeeBalances(ctx, employeeID)
}

var _ emplUc.BalanceAssigner = (*BalanceAssignerAdapter)(nil)
