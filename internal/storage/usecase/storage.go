package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	errors "hrms/internal/pkg/apperror"
	"hrms/internal/storage/entity"
	"hrms/internal/storage/repository"
)

type StorageUsecase struct {
	docRepo repository.DocumentRepository
	storage repository.ObjectStorage
	log     *logrus.Logger
	maxSize int64
}

func NewStorageUsecase(
	docRepo repository.DocumentRepository,
	storage repository.ObjectStorage,
	log *logrus.Logger,
	maxSize int64,
) *StorageUsecase {
	return &StorageUsecase{
		docRepo: docRepo,
		storage: storage,
		log:     log,
		maxSize: maxSize,
	}
}

func (uc *StorageUsecase) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, uploadedBy, module string, referenceID *string) (*entity.Document, error) {
	if header.Size > uc.maxSize {
		return nil, errors.NewInvalidInput(fmt.Sprintf("file too large: max %d bytes", uc.maxSize))
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.NewInternal("failed to read file")
	}
	defer file.Close()

	mimeType := entity.DetectMimeType(data)
	if mimeType == "" || !entity.IsAllowedMimeType(mimeType) {
		return nil, errors.NewInvalidInput("unsupported file type")
	}

	ext := entity.ExtensionFromMime(mimeType)
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(header.Filename))
	}

	now := time.Now()
	storageKey := fmt.Sprintf("%s/%04d/%02d/%s%s", module, now.Year(), now.Month(), uuid.New().String(), ext)

	if err := uc.storage.Upload(ctx, storageKey, bytes.NewReader(data), header.Size, mimeType); err != nil {
		uc.log.WithError(err).Error("failed to upload to R2")
		return nil, errors.NewInternal("failed to store file")
	}

	doc := entity.NewDocument(header.Filename, mimeType, header.Size, &uploadedBy, module, referenceID)
	doc.StorageKey = storageKey

	if err := uc.docRepo.Create(ctx, doc); err != nil {
		uc.log.WithError(err).Error("failed to save document metadata")
		if delErr := uc.storage.Delete(ctx, storageKey); delErr != nil {
			uc.log.WithError(delErr).Error("failed to rollback R2 upload")
		}
		return nil, errors.NewInternal("failed to save file metadata")
	}

	return doc, nil
}

func (uc *StorageUsecase) Download(ctx context.Context, docID string) (io.ReadCloser, *entity.Document, error) {
	doc, err := uc.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, nil, errors.NewNotFound("document not found")
	}

	reader, err := uc.storage.Download(ctx, doc.StorageKey)
	if err != nil {
		uc.log.WithError(err).Error("failed to download from R2")
		return nil, nil, errors.NewInternal("failed to retrieve file")
	}

	return reader, doc, nil
}

func (uc *StorageUsecase) Delete(ctx context.Context, docID string) error {
	doc, err := uc.docRepo.FindByID(ctx, docID)
	if err != nil {
		return errors.NewNotFound("document not found")
	}

	if err := uc.storage.Delete(ctx, doc.StorageKey); err != nil {
		uc.log.WithError(err).Error("failed to delete from R2")
		return errors.NewInternal("failed to delete file")
	}

	return uc.docRepo.DeleteSoft(ctx, docID)
}

func (uc *StorageUsecase) Replace(ctx context.Context, file multipart.File, header *multipart.FileHeader, uploadedBy, module string, referenceID *string) (*entity.Document, error) {
	if referenceID == nil || *referenceID == "" {
		return nil, errors.NewInvalidInput("reference_id is required for replace")
	}

	if header.Size > uc.maxSize {
		return nil, errors.NewInvalidInput(fmt.Sprintf("file too large: max %d bytes", uc.maxSize))
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.NewInternal("failed to read file")
	}
	defer file.Close()

	mimeType := entity.DetectMimeType(data)
	if mimeType == "" || !entity.IsAllowedMimeType(mimeType) {
		return nil, errors.NewInvalidInput("unsupported file type")
	}

	ext := entity.ExtensionFromMime(mimeType)
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(header.Filename))
	}

	now := time.Now()
	storageKey := fmt.Sprintf("%s/%04d/%02d/%s%s", module, now.Year(), now.Month(), uuid.New().String(), ext)

	if err := uc.storage.Upload(ctx, storageKey, bytes.NewReader(data), header.Size, mimeType); err != nil {
		uc.log.WithError(err).Error("failed to upload to R2")
		return nil, errors.NewInternal("failed to store file")
	}

	doc := entity.NewDocument(header.Filename, mimeType, header.Size, &uploadedBy, module, referenceID)
	doc.StorageKey = storageKey

	if err := uc.docRepo.Create(ctx, doc); err != nil {
		uc.log.WithError(err).Error("failed to save document metadata")
		if delErr := uc.storage.Delete(ctx, storageKey); delErr != nil {
			uc.log.WithError(delErr).Error("failed to rollback R2 upload")
		}
		return nil, errors.NewInternal("failed to save file metadata")
	}

	oldDocs, listErr := uc.docRepo.ListByModule(ctx, module, *referenceID)
	if listErr == nil {
		for i := range oldDocs {
			if oldDocs[i].ID == doc.ID {
				continue
			}
			if delErr := uc.storage.Delete(ctx, oldDocs[i].StorageKey); delErr != nil {
				uc.log.WithError(delErr).WithField("doc_id", oldDocs[i].ID).Error("failed to delete old file from R2")
			}
			if softErr := uc.docRepo.DeleteSoft(ctx, oldDocs[i].ID); softErr != nil {
				uc.log.WithError(softErr).WithField("doc_id", oldDocs[i].ID).Error("failed to soft-delete old doc")
			}
		}
	} else {
		uc.log.WithError(listErr).Warn("failed to list old docs for cleanup (will be handled by GC)")
	}

	return doc, nil
}

func (uc *StorageUsecase) ListByModule(ctx context.Context, module, referenceID string) ([]entity.Document, error) {
	docs, err := uc.docRepo.ListByModule(ctx, module, referenceID)
	if err != nil {
		return nil, errors.NewInternal("failed to list documents")
	}
	return docs, nil
}
