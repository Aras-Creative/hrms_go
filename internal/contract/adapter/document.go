package adapter

import (
	"context"
	"fmt"
	"io"

	contractUc "hrms/internal/contract/usecase"
	storageRepo "hrms/internal/storage/repository"
)

// ObjectStorageAdapter wraps storage.ObjectStorage.
type ObjectStorageAdapter struct {
	store storageRepo.ObjectStorage
}

func NewObjectStorageAdapter(store storageRepo.ObjectStorage) *ObjectStorageAdapter {
	return &ObjectStorageAdapter{store: store}
}

func (a *ObjectStorageAdapter) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	return a.store.Upload(ctx, key, reader, size, contentType)
}

func (a *ObjectStorageAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return a.store.Download(ctx, key)
}

func (a *ObjectStorageAdapter) Delete(ctx context.Context, key string) error {
	return a.store.Delete(ctx, key)
}

var _ contractUc.ObjectStorage = (*ObjectStorageAdapter)(nil)

// DocumentMetadataAdapter wraps storage.DocumentRepository to satisfy
// the contract usecase's DocumentMetadataRepository interface.
type DocumentMetadataAdapter struct {
	repo storageRepo.DocumentRepository
}

func NewDocumentMetadataAdapter(repo storageRepo.DocumentRepository) *DocumentMetadataAdapter {
	return &DocumentMetadataAdapter{repo: repo}
}

func (a *DocumentMetadataAdapter) CreateForModule(ctx context.Context, originalName, mimeType string, uploaderID *string, module, referenceID string, sizeBytes int64, id, storageKey string) (string, error) {
	err := a.repo.CreateForModule(ctx, originalName, mimeType, uploaderID, module, referenceID, sizeBytes, id, storageKey)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (a *DocumentMetadataAdapter) FindByID(ctx context.Context, id string) (originalName, mimeType, storageKey string, sizeBytes int64, err error) {
	doc, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return "", "", "", 0, err
	}
	if doc == nil {
		return "", "", "", 0, fmt.Errorf("document not found: %s", id)
	}
	return doc.OriginalName, doc.MimeType, doc.StorageKey, doc.SizeBytes, nil
}

var _ contractUc.DocumentMetadataRepository = (*DocumentMetadataAdapter)(nil)
