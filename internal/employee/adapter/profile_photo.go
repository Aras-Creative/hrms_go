package adapter

import (
	"context"

	storageRepo "hrms/internal/storage/repository"
	emplUc "hrms/internal/employee/usecase"
)

type ProfilePhotoResolverAdapter struct {
	docRepo  storageRepo.DocumentRepository
	resolver storageRepo.URLResolver
}

func NewProfilePhotoResolverAdapter(docRepo storageRepo.DocumentRepository, resolver storageRepo.URLResolver) *ProfilePhotoResolverAdapter {
	return &ProfilePhotoResolverAdapter{docRepo: docRepo, resolver: resolver}
}

func (a *ProfilePhotoResolverAdapter) ResolveURL(ctx context.Context, documentID string) (string, error) {
	doc, err := a.docRepo.FindByID(ctx, documentID)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", nil
	}
	return a.resolver.PublicURL(doc.StorageKey), nil
}

func (a *ProfilePhotoResolverAdapter) ResolveURLs(ctx context.Context, documentIDs []string) (map[string]string, error) {
	docs, err := a.docRepo.FindByIDs(ctx, documentIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(docs))
	for id, doc := range docs {
		result[id] = a.resolver.PublicURL(doc.StorageKey)
	}
	return result, nil
}

var _ emplUc.ProfilePhotoResolver = (*ProfilePhotoResolverAdapter)(nil)
