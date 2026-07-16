package adapter

import (
	"context"

	designationRepo "hrms/internal/designation/repository"
	emplUc "hrms/internal/employee/usecase"
)

type DesignationFetcherAdapter struct {
	repo designationRepo.DesignationRepository
}

func NewDesignationFetcherAdapter(repo designationRepo.DesignationRepository) *DesignationFetcherAdapter {
	return &DesignationFetcherAdapter{repo: repo}
}

func (a *DesignationFetcherAdapter) FindCodeByID(ctx context.Context, id string) (string, error) {
	d, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return d.Code, nil
}

var _ emplUc.DesignationFetcher = (*DesignationFetcherAdapter)(nil)
