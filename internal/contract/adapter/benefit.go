package adapter

import (
	"context"
	"fmt"

	contractUc "hrms/internal/contract/usecase"
	payrollRepo "hrms/internal/payroll/repository"
)

type BenefitFetcherAdapter struct {
	empBenefitRepo  payrollRepo.EmployeeBenefitRepository
	benefitTypeRepo payrollRepo.BenefitTypeRepository
}

func NewBenefitFetcherAdapter(empBenefitRepo payrollRepo.EmployeeBenefitRepository, benefitTypeRepo payrollRepo.BenefitTypeRepository) *BenefitFetcherAdapter {
	return &BenefitFetcherAdapter{empBenefitRepo: empBenefitRepo, benefitTypeRepo: benefitTypeRepo}
}

func (a *BenefitFetcherAdapter) FindByEmployeeID(ctx context.Context, employeeID string) ([]contractUc.BenefitItem, error) {
	empBenefits, err := a.empBenefitRepo.FindByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee benefits: %w", err)
	}

	typeIDs := make(map[string]struct{})
	for _, eb := range empBenefits {
		{
			typeIDs[eb.BenefitTypeID] = struct{}{}
		}
	}

	items := make([]contractUc.BenefitItem, 0, len(typeIDs))
	for typeID := range typeIDs {
		bt, err := a.benefitTypeRepo.FindByID(ctx, typeID)
		if err != nil || bt == nil {
			continue
		}
		items = append(items, contractUc.BenefitItem{
			Name: bt.Name,
		})
	}
	return items, nil
}

var _ contractUc.BenefitFetcher = (*BenefitFetcherAdapter)(nil)
