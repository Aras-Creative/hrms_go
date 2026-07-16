package repository

import (
	"context"
	"io"

	"hrms/internal/storage/entity"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *entity.Document) error
	CreateForModule(ctx context.Context, originalName, mimeType string, uploaderID *string, module, referenceID string, sizeBytes int64, id, storageKey string) error
	FindByID(ctx context.Context, id string) (*entity.Document, error)
	FindByIDs(ctx context.Context, ids []string) (map[string]*entity.Document, error)
	ListByModule(ctx context.Context, module, referenceID string) ([]entity.Document, error)
	DeleteSoft(ctx context.Context, id string) error
}

type ObjectStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type URLResolver interface {
	PublicURL(storageKey string) string
}
