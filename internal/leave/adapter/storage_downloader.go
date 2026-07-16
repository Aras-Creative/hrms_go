package adapter

import (
	"context"

	storageRepo "hrms/internal/storage/repository"
)

type StorageAttachmentResolver struct {
	docRepo     storageRepo.DocumentRepository
	urlResolver storageRepo.URLResolver
}

func NewStorageAttachmentResolver(docRepo storageRepo.DocumentRepository, urlResolver storageRepo.URLResolver) *StorageAttachmentResolver {
	return &StorageAttachmentResolver{docRepo: docRepo, urlResolver: urlResolver}
}

func (a *StorageAttachmentResolver) Resolve(ctx context.Context, docID string) (string, error) {
	doc, err := a.docRepo.FindByID(ctx, docID)
	if err != nil {
		return "", err
	}
	return a.urlResolver.PublicURL(doc.StorageKey), nil
}
