package entity

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var AllowedMIMETypes = map[string]string{
	"application/pdf": ".pdf",
	"text/csv":        ".csv",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       ".xlsx",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type Document struct {
	ID           string
	OriginalName string
	MimeType     string
	SizeBytes    int64
	StorageKey   string
	UploadedBy   *string
	Module       string
	ReferenceID  *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func NewDocument(originalName, mimeType string, sizeBytes int64, uploadedBy *string, module string, referenceID *string) *Document {
	now := time.Now()
	return &Document{
		ID:           "",
		OriginalName: originalName,
		MimeType:     mimeType,
		SizeBytes:    sizeBytes,
		UploadedBy:   uploadedBy,
		Module:       module,
		ReferenceID:  referenceID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func ReconstituteDocument(
	id, originalName, mimeType string,
	sizeBytes int64,
	storageKey string,
	uploadedBy *string,
	module string,
	referenceID *string,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Document {
	return &Document{
		ID:           id,
		OriginalName: originalName,
		MimeType:     mimeType,
		SizeBytes:    sizeBytes,
		StorageKey:   storageKey,
		UploadedBy:   uploadedBy,
		Module:       module,
		ReferenceID:  referenceID,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DeletedAt:    deletedAt,
	}
}

func DetectMimeType(data []byte) string {
	detected := http.DetectContentType(data)
	clean := strings.Split(detected, ";")[0]
	clean = strings.ToLower(strings.TrimSpace(clean))

	switch clean {
	case "application/pdf", "text/csv", "text/plain":
		return "text/csv"
	case "application/vnd.ms-excel":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "application/msword":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	if _, ok := AllowedMIMETypes[clean]; ok {
		return clean
	}

	return ""
}

func IsAllowedMimeType(mimeType string) bool {
	_, ok := AllowedMIMETypes[mimeType]
	return ok
}

func ExtensionFromMime(mimeType string) string {
	return AllowedMIMETypes[mimeType]
}

func ExtensionFromFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	for _, allowed := range AllowedMIMETypes {
		if allowed == ext {
			return ext
		}
	}
	return ""
}

func MimeTypeFromFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return ""
	}
	return strings.Split(mimeType, ";")[0]
}
