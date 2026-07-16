package adapter

import (
	"context"

	storageRepo "hrms/internal/storage/repository"
	"hrms/internal/setting/usecase"
)

type LogoResolverAdapter struct {
	docRepo  storageRepo.DocumentRepository
	resolver storageRepo.URLResolver
}

func NewLogoResolverAdapter(docRepo storageRepo.DocumentRepository, resolver storageRepo.URLResolver) *LogoResolverAdapter {
	return &LogoResolverAdapter{docRepo: docRepo, resolver: resolver}
}

func (a *LogoResolverAdapter) ResolveURL(ctx context.Context, documentID string) (string, error) {
	doc, err := a.docRepo.FindByID(ctx, documentID)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", nil
	}
	return a.resolver.PublicURL(doc.StorageKey), nil
}

var _ usecase.LogoResolver = (*LogoResolverAdapter)(nil)
