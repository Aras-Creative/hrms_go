package adapter

import (
	"context"
	"fmt"

	payrollRepo "hrms/internal/payroll/repository"
	contractUc "hrms/internal/contract/usecase"
)

type SalaryFetcherAdapter struct {
	repo payrollRepo.EmployeeBaseSalaryRepository
}

func NewSalaryFetcherAdapter(repo payrollRepo.EmployeeBaseSalaryRepository) *SalaryFetcherAdapter {
	return &SalaryFetcherAdapter{repo: repo}
}

func (a *SalaryFetcherAdapter) FindCurrentByEmployeeIDs(ctx context.Context, employeeIDs []string) (map[string]string, error) {
	salaries, err := a.repo.FindCurrentByEmployeeIDs(ctx, employeeIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(salaries))
	for empID, s := range salaries {
		result[empID] = fmt.Sprintf("%.2f", s.Amount.Float())
	}
	return result, nil
}

var _ contractUc.SalaryFetcher = (*SalaryFetcherAdapter)(nil)
