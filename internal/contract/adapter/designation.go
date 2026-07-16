package adapter

import (
	"context"

	designationRepo "hrms/internal/designation/repository"
	contractUc "hrms/internal/contract/usecase"
)

type DesignationFetcherAdapter struct {
	repo designationRepo.DesignationRepository
}

func NewDesignationFetcherAdapter(repo designationRepo.DesignationRepository) *DesignationFetcherAdapter {
	return &DesignationFetcherAdapter{repo: repo}
}

func (a *DesignationFetcherAdapter) FindNamesByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	designations, err := a.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(designations))
	for _, d := range designations {
		result[d.ID] = d.Name
	}
	return result, nil
}

var _ contractUc.DesignationFetcher = (*DesignationFetcherAdapter)(nil)
