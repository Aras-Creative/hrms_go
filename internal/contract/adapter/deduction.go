package adapter

import (
	"context"
	"fmt"

	payrollRepo "hrms/internal/payroll/repository"
	contractUc "hrms/internal/contract/usecase"
)

type DeductionFetcherAdapter struct {
	empDeductionRepo  payrollRepo.EmployeeDeductionRepository
	deductionTypeRepo payrollRepo.DeductionTypeRepository
}

func NewDeductionFetcherAdapter(empDeductionRepo payrollRepo.EmployeeDeductionRepository, deductionTypeRepo payrollRepo.DeductionTypeRepository) *DeductionFetcherAdapter {
	return &DeductionFetcherAdapter{empDeductionRepo: empDeductionRepo, deductionTypeRepo: deductionTypeRepo}
}

func (a *DeductionFetcherAdapter) FindByEmployeeID(ctx context.Context, employeeID string) ([]contractUc.DeductionItem, error) {
	empDeds, err := a.empDeductionRepo.FindByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee deductions: %w", err)
	}

	typeIDs := make(map[string]struct{})
	for _, ed := range empDeds {
		if ed.EndDate == nil {
			typeIDs[ed.DeductionTypeID] = struct{}{}
		}
	}

	items := make([]contractUc.DeductionItem, 0, len(typeIDs))
	for typeID := range typeIDs {
		dt, err := a.deductionTypeRepo.FindByID(ctx, typeID)
		if err != nil || dt == nil {
			continue
		}
		items = append(items, contractUc.DeductionItem{
			Name: dt.Name,
		})
	}
	return items, nil
}

var _ contractUc.DeductionFetcher = (*DeductionFetcherAdapter)(nil)
