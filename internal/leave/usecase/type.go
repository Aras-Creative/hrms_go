package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
	errors "hrms/internal/pkg/apperror"
)

func (uc *LeaveUsecase) CreateLeaveType(ctx context.Context, input models.CreateLeaveTypeInput) (*entity.LeaveType, error) {
	lt := entity.NewLeaveType(input.Name, input.DefaultDays, input.IsPaid, input.IsUnlimited, input.IsHalfDay)
	if err := uc.leaveTypeRepo.Create(ctx, lt); err != nil {
		return nil, fmt.Errorf("failed to create leave type: %w", err)
	}

	if !input.IsUnlimited && input.DefaultDays > 0 {
		if err := uc.SeedBalancesForLeaveType(ctx, lt.ID, input.DefaultDays); err != nil {
			return nil, fmt.Errorf("failed to seed leave balances: %w", err)
		}
	}

	return lt, nil
}

func (uc *LeaveUsecase) GetLeaveTypeByID(ctx context.Context, id string) (*entity.LeaveType, error) {
	lt, err := uc.leaveTypeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find leave type: %w", err)
	}
	if lt == nil {
		return nil, errors.NewNotFound("leave type not found")
	}
	return lt, nil
}

func (uc *LeaveUsecase) GetAllLeaveTypes(ctx context.Context) ([]*entity.LeaveType, error) {
	return uc.leaveTypeRepo.FindAllActive(ctx)
}

func (uc *LeaveUsecase) UpdateLeaveType(ctx context.Context, id string, input models.UpdateLeaveTypeInput) (*entity.LeaveType, error) {
	lt, err := uc.leaveTypeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find leave type: %w", err)
	}
	if lt == nil {
		return nil, errors.NewNotFound("leave type not found")
	}

	if input.Name != "" {
		lt.Rename(input.Name)
	}
	if input.DefaultDays >= 0 {
		lt.SetDefaultDays(input.DefaultDays)
	}
	if input.IsPaid != lt.IsPaid {
		lt.SetPaidStatus(input.IsPaid)
	}
	if input.IsUnlimited != lt.IsUnlimited {
		lt.SetUnlimited(input.IsUnlimited)
	}
	if input.IsHalfDay != lt.IsHalfDay {
		lt.IsHalfDay = input.IsHalfDay
		lt.UpdatedAt = time.Now()
	}

	if err := uc.leaveTypeRepo.Update(ctx, lt); err != nil {
		return nil, fmt.Errorf("failed to update leave type: %w", err)
	}
	return lt, nil
}

func (uc *LeaveUsecase) DeleteLeaveType(ctx context.Context, id string) error {
	lt, err := uc.leaveTypeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find leave type: %w", err)
	}
	if lt == nil {
		return errors.NewNotFound("leave type not found")
	}
	lt.Disable()
	return uc.leaveTypeRepo.Update(ctx, lt)
}

func (uc *LeaveUsecase) SeedBalancesForLeaveType(ctx context.Context, leaveTypeID string, defaultDays int) error {
	employeeIDs, err := uc.employeeFetcher.GetAllActiveIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch employee IDs: %w", err)
	}

	year := time.Now().Year()
	for _, empID := range employeeIDs {
		existing, err := uc.leaveBalanceRepo.FindByEmployeeAndTypeYear(ctx, empID, leaveTypeID, year)
		if err != nil {
			return fmt.Errorf("failed to check existing balance: %w", err)
		}
		if existing != nil {
			continue
		}
		lb := entity.NewLeaveBalance(empID, leaveTypeID, year, defaultDays)
		if err := uc.leaveBalanceRepo.Create(ctx, lb); err != nil {
			return fmt.Errorf("failed to create balance for employee %s: %w", empID, err)
		}
	}
	return nil
}

func (uc *LeaveUsecase) SeedBalancesForLeaveTypeAsync(leaveTypeID string, defaultDays int) {
	go func() {
		ctx := context.Background()
		if err := uc.SeedBalancesForLeaveType(ctx, leaveTypeID, defaultDays); err != nil {
			uc.log.WithField("leave_type_id", leaveTypeID).WithField("error", err).Error("failed to seed leave balances")
		}
	}()
}

func (uc *LeaveUsecase) AssignEmployeeBalances(ctx context.Context, employeeID string) error {
	leaveTypes, err := uc.leaveTypeRepo.FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active leave types: %w", err)
	}

	year := time.Now().Year()
	for _, lt := range leaveTypes {
		if lt.IsUnlimited || lt.DefaultDays <= 0 {
			continue
		}

		existing, err := uc.leaveBalanceRepo.FindByEmployeeAndTypeYear(ctx, employeeID, lt.ID, year)
		if err != nil {
			return fmt.Errorf("failed to check existing balance: %w", err)
		}
		if existing != nil {
			continue
		}

		lb := entity.NewLeaveBalance(employeeID, lt.ID, year, lt.DefaultDays)
		if err := uc.leaveBalanceRepo.Create(ctx, lb); err != nil {
			return fmt.Errorf("failed to create balance: %w", err)
		}
	}

	return nil
}
