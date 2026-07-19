package usecase

import (
	"context"
	"time"

	contractEntity "hrms/internal/contract/entity"
	errors "hrms/internal/pkg/apperror"
)

type ContractFinder interface {
	FindContractByID(ctx context.Context, id string) (*contractEntity.Contract, error)
	UpdateContract(ctx context.Context, c *contractEntity.Contract) error
}

type EmployeeTerminator interface {
	FindEmployeeUserID(ctx context.Context, employeeID string) (string, error)
	TerminateEmployee(ctx context.Context, employeeID string, terminationDate time.Time) error
}

type UserDeactivator interface {
	Deactivate(ctx context.Context, userID string) error
}

type SessionRevoker interface {
	RevokeAllByUserID(ctx context.Context, userID string) error
}

type DeviceRevoker interface {
	RevokeDeviceByUserID(ctx context.Context, userID string) error
}

type WorkPatternDeactivator interface {
	DeactivateCurrent(ctx context.Context, employeeID string, validTo time.Time) error
}

type ScheduleOverrideDeleter interface {
	DeleteFutureOverridesByEmployee(ctx context.Context, employeeID string, since time.Time) error
}

type TerminationUsecase struct {
	contractRepo ContractFinder
	empTerminator EmployeeTerminator
	userDeactiv  UserDeactivator
	sessionRev   SessionRevoker
	deviceRev    DeviceRevoker
	ewpDeactiv   WorkPatternDeactivator
	overrideDel  ScheduleOverrideDeleter
}

func NewTerminationUsecase(
	contractRepo ContractFinder,
	empTerminator EmployeeTerminator,
	userDeactiv UserDeactivator,
	sessionRev SessionRevoker,
	deviceRev DeviceRevoker,
	ewpDeactiv WorkPatternDeactivator,
	overrideDel ScheduleOverrideDeleter,
) *TerminationUsecase {
	return &TerminationUsecase{
		contractRepo:  contractRepo,
		empTerminator: empTerminator,
		userDeactiv:   userDeactiv,
		sessionRev:    sessionRev,
		deviceRev:     deviceRev,
		ewpDeactiv:    ewpDeactiv,
		overrideDel:   overrideDel,
	}
}

type TerminateContractInput struct {
	ContractID      string
	TerminationDate time.Time
}

func (uc *TerminationUsecase) TerminateContract(ctx context.Context, input TerminateContractInput) error {
	c, err := uc.contractRepo.FindContractByID(ctx, input.ContractID)
	if err != nil {
		return errors.WrapInternal("find contract", err)
	}
	if c == nil {
		return errors.NewNotFound("contract not found")
	}

	if err := c.Terminate(); err != nil {
		return errors.NewInvalidInput(err.Error())
	}

	// 1. Terminate contract
	if err := uc.contractRepo.UpdateContract(ctx, c); err != nil {
		return errors.WrapInternal("update contract", err)
	}

	// 2. Terminate employee
	if err := uc.empTerminator.TerminateEmployee(ctx, c.EmployeeID, input.TerminationDate); err != nil {
		return errors.WrapInternal("terminate employee", err)
	}

	// 3. Deactivate user + revoke sessions + devices (best-effort)
	empUserID, err := uc.empTerminator.FindEmployeeUserID(ctx, c.EmployeeID)
	if err != nil {
		return errors.WrapInternal("find employee user", err)
	}
	if empUserID != "" {
		_ = uc.userDeactiv.Deactivate(ctx, empUserID)
		_ = uc.sessionRev.RevokeAllByUserID(ctx, empUserID)
		_ = uc.deviceRev.RevokeDeviceByUserID(ctx, empUserID)
	}

	// 4. Deactivate work pattern (set valid_to = termination date)
	if err := uc.ewpDeactiv.DeactivateCurrent(ctx, c.EmployeeID, input.TerminationDate); err != nil {
		return errors.WrapInternal("deactivate work pattern", err)
	}

	// 5. Delete future schedule overrides from termination date onward
	if err := uc.overrideDel.DeleteFutureOverridesByEmployee(ctx, c.EmployeeID, input.TerminationDate); err != nil {
		return errors.WrapInternal("delete future overrides", err)
	}

	return nil
}
