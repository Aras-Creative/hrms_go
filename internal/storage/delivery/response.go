package delivery

import (
	"time"

	"hrms/internal/storage/entity"
	"hrms/internal/storage/repository"
)

type DocumentResponse struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	Module       string    `json:"module"`
	ReferenceID  *string   `json:"reference_id,omitempty"`
	UploadedBy   *string   `json:"uploaded_by"`
	URL          string    `json:"url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func newDocumentResponse(d *entity.Document, resolvers ...repository.URLResolver) DocumentResponse {
	var url string
	for _, r := range resolvers {
		url = r.PublicURL(d.StorageKey)
		if url != "" {
			break
		}
	}

	return DocumentResponse{
		ID:           d.ID,
		OriginalName: d.OriginalName,
		MimeType:     d.MimeType,
		SizeBytes:    d.SizeBytes,
		Module:       d.Module,
		ReferenceID:  d.ReferenceID,
		UploadedBy:   d.UploadedBy,
		URL:          url,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}


