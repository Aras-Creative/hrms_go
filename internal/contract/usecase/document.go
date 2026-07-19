package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/repository"
	errors "hrms/internal/pkg/apperror"
)

type ContractDocumentMetadata struct {
	MimeType     string
	OriginalName string
	SizeBytes    int64
}

type ObjectStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type DocumentMetadataRepository interface {
	CreateForModule(ctx context.Context, originalName, mimeType string, uploaderID *string, module, referenceID string, sizeBytes int64, id, storageKey string) (string, error)
	FindByID(ctx context.Context, id string) (originalName, mimeType, storageKey string, sizeBytes int64, err error)
}

type DocumentUsecase struct {
	renderUC    *RenderUsecase
	objStore    ObjectStorage
	storageDoc  DocumentMetadataRepository
	contractDoc repository.DocumentRepository
}

func NewDocumentUsecase(
	renderUC *RenderUsecase,
	objStore ObjectStorage,
	storageDoc DocumentMetadataRepository,
	contractDoc repository.DocumentRepository,
) *DocumentUsecase {
	return &DocumentUsecase{
		renderUC:    renderUC,
		objStore:    objStore,
		storageDoc:  storageDoc,
		contractDoc: contractDoc,
	}
}

func (uc *DocumentUsecase) StorePDF(ctx context.Context, contractID, contractNumber, signatoryName, signatoryTitle string) (documentID, contentHash string, err error) {
	pdfBytes, err := uc.renderUC.Preview(ctx, contractID, signatoryName, signatoryTitle)
	if err != nil {
		return "", "", errors.WrapInternal("generate pdf", err)
	}
	return uc.storePDFBytes(ctx, contractID, contractNumber, pdfBytes)
}

func (uc *DocumentUsecase) StorePDFWithSignings(ctx context.Context, contractID, contractNumber, signatoryName, signatoryTitle string, signings []*entity.ContractSigning) (documentID, contentHash string, err error) {
	pdfBytes, err := uc.renderUC.PreviewWithSignings(ctx, contractID, signatoryName, signatoryTitle, signings)
	if err != nil {
		return "", "", errors.WrapInternal("generate pdf", err)
	}
	return uc.storePDFBytes(ctx, contractID, contractNumber, pdfBytes)
}

func (uc *DocumentUsecase) storePDFBytes(ctx context.Context, contractID, contractNumber string, pdfBytes []byte) (documentID, contentHash string, err error) {

	hash := sha256.Sum256(pdfBytes)
	contentHash = hex.EncodeToString(hash[:])

	docID := uuid.New().String()
	now := time.Now()
	storageKey := fmt.Sprintf("contracts/%04d/%02d/%s.pdf", now.Year(), now.Month(), docID)

	if err := uc.objStore.Upload(ctx, storageKey, bytes.NewReader(pdfBytes), int64(len(pdfBytes)), "application/pdf"); err != nil {
		return "", "", errors.WrapInternal("upload pdf", err)
	}

	originalName := fmt.Sprintf("Perjanjian Kerja - %s.pdf", contractNumber)
	if _, err := uc.storageDoc.CreateForModule(ctx, originalName, "application/pdf", nil, "contracts", contractID, int64(len(pdfBytes)), docID, storageKey); err != nil {
		if delErr := uc.objStore.Delete(ctx, storageKey); delErr != nil {
		}
		return "", "", errors.WrapInternal("create document record", err)
	}

	contractDoc := entity.NewContractDocument(contractID, docID, contentHash)
	if err := uc.contractDoc.CreateContractDocument(ctx, contractDoc); err != nil {
		return "", "", errors.WrapInternal("link contract document", err)
	}

	return docID, contentHash, nil
}

func (uc *DocumentUsecase) DownloadPDF(ctx context.Context, contractID string) (io.ReadCloser, *ContractDocumentMetadata, error) {
	contractDoc, err := uc.contractDoc.FindContractDocumentByContractID(ctx, contractID)
	if err != nil {
		return nil, nil, errors.WrapInternal("find contract document", err)
	}
	if contractDoc == nil {
		return nil, nil, errors.NewInvalidInput("contract document not yet generated")
	}

	originalName, mimeType, storageKey, sizeBytes, err := uc.storageDoc.FindByID(ctx, contractDoc.DocumentID)
	if err != nil {
		return nil, nil, errors.WrapInternal("find document metadata", err)
	}

	reader, err := uc.objStore.Download(ctx, storageKey)
	if err != nil {
		return nil, nil, errors.WrapInternal("download object", err)
	}

	return reader, &ContractDocumentMetadata{
		MimeType:     mimeType,
		OriginalName: originalName,
		SizeBytes:    sizeBytes,
	}, nil
}
