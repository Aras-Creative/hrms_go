package adapter

import (
	"context"

	contractRepo "hrms/internal/contract/repository"
	emplUc "hrms/internal/employee/usecase"
)

type CurrentContractFetcherAdapter struct {
	repo contractRepo.ContractRepository
}

func NewCurrentContractFetcherAdapter(repo contractRepo.ContractRepository) *CurrentContractFetcherAdapter {
	return &CurrentContractFetcherAdapter{repo: repo}
}

func (a *CurrentContractFetcherAdapter) FindCurrentByEmployeeID(ctx context.Context, employeeID string) (*emplUc.ContractBrief, error) {
	c, err := a.repo.FindCurrentByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	return &emplUc.ContractBrief{
		ContractType: c.ContractType,
		StartDate:    c.StartDate,
		EndDate:      c.EndDate,
	}, nil
}

var _ emplUc.CurrentContractFetcher = (*CurrentContractFetcherAdapter)(nil)
