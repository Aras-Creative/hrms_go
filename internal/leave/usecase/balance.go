package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
	"hrms/internal/leave/repository"
	errors "hrms/internal/pkg/apperror"
)

func (uc *LeaveUsecase) ListBalances(ctx context.Context, input models.ListBalanceInput) (*models.ListBalanceResult, error) {
	rows, total, err := uc.leaveBalanceRepo.FindAll(ctx, repository.BalanceFilter{
		LeaveTypeID: input.LeaveTypeID,
		Search:      input.Search,
		Year:        input.Year,
		Page:        input.Page,
		PerPage:     input.PerPage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list balances: %w", err)
	}
	items := make([]models.LeaveBalance, 0, len(rows))
	for _, r := range rows {
		items = append(items, *r)
	}
	return &models.ListBalanceResult{Items: items, Total: total}, nil
}

func (uc *LeaveUsecase) UpdateBalance(ctx context.Context, input models.UpdateLeaveBalanceInput) (*entity.LeaveBalance, error) {
	b, err := uc.leaveBalanceRepo.FindByEmployeeAndTypeYear(ctx, input.EmployeeID, input.LeaveTypeID, input.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to find balance: %w", err)
	}
	if b == nil {
		return nil, errors.NewNotFound("leave balance not found")
	}

	b.SetTotalDays(input.TotalDays)

	if err := uc.leaveBalanceRepo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}
	return b, nil
}

func (uc *LeaveUsecase) GetEmployeeBalance(ctx context.Context, employeeID, leaveTypeID string) (*models.LeaveBalance, error) {
	year := time.Now().Year()
	b, err := uc.leaveBalanceRepo.FindByEmployeeAndTypeYear(ctx, employeeID, leaveTypeID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to find balance: %w", err)
	}
	if b == nil {
		return nil, errors.NewNotFound("leave balance not found")
	}

	lt, err := uc.leaveTypeRepo.FindByID(ctx, leaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find leave type: %w", err)
	}

	m := models.LeaveBalance{
		ID:          b.ID,
		EmployeeID:  b.EmployeeID,
		LeaveTypeID: b.LeaveTypeID,
		Year:        b.Year,
		TotalDays:   b.TotalDays,
		UsedDays:    b.UsedDays,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
	if lt != nil {
		m.LeaveTypeName = lt.Name
	}
	return &m, nil
}
