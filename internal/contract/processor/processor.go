package processor

import (
	"context"
	"log/slog"
	"time"

	"hrms/internal/contract/entity"
	contractRepo "hrms/internal/contract/repository"
	emplEntity "hrms/internal/employee/entity"
)

type EmployeeStatusUpdater interface {
	FindByID(ctx context.Context, id string) (*emplEntity.Employee, error)
	Update(ctx context.Context, e *emplEntity.Employee) error
}

type ExpiryProcessor struct {
	contractRepo contractRepo.ContractRepository
	empUpdater   EmployeeStatusUpdater
}

func NewExpiryProcessor(contractRepo contractRepo.ContractRepository, empUpdater EmployeeStatusUpdater) *ExpiryProcessor {
	return &ExpiryProcessor{
		contractRepo: contractRepo,
		empUpdater:   empUpdater,
	}
}

func (p *ExpiryProcessor) Process(ctx context.Context, asOf time.Time) error {
	contracts, err := p.contractRepo.FindActiveContractsPastEndDate(ctx, asOf)
	if err != nil {
		return err
	}

	expired := 0
	for _, c := range contracts {
		if err := p.expireOne(ctx, c); err != nil {
			slog.Error("contract expiry: failed",
				"contract_id", c.ID, "employee_id", c.EmployeeID, "error", err)
			continue
		}
		expired++
	}

	if expired > 0 {
		slog.Info("contract expiry: processed", "expired", expired, "as_of", asOf.Format("2006-01-02"))
	}
	return nil
}

func (p *ExpiryProcessor) expireOne(ctx context.Context, c *entity.Contract) error {
	if err := c.Expire(); err != nil {
		return err
	}
	if err := p.contractRepo.UpdateContract(ctx, c); err != nil {
		return err
	}

	hasOther, err := p.contractRepo.HasOtherActiveContract(ctx, c.EmployeeID, c.ID)
	if err != nil {
		return err
	}
	if hasOther {
		return nil
	}

	emp, err := p.empUpdater.FindByID(ctx, c.EmployeeID)
	if err != nil {
		return err
	}
	if emp == nil {
		return nil
	}
	emp.Status = emplEntity.StatusExpiredContract
	emp.IsActive = false
	now := time.Now()
	emp.UpdatedAt = now
	return p.empUpdater.Update(ctx, emp)
}
