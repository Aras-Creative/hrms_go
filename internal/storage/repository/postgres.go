package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"hrms/internal/storage/entity"
)

type documentModel struct {
	ID           string     `db:"id"`
	OriginalName string     `db:"original_name"`
	MimeType     string     `db:"mime_type"`
	SizeBytes    int64      `db:"size_bytes"`
	StorageKey   string     `db:"storage_key"`
	UploadedBy   *string    `db:"uploaded_by"`
	Module       string     `db:"module"`
	ReferenceID  *string    `db:"reference_id"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

func (m *documentModel) toEntity() *entity.Document {
	return entity.ReconstituteDocument(
		m.ID, m.OriginalName, m.MimeType,
		m.SizeBytes, m.StorageKey,
		m.UploadedBy, m.Module, m.ReferenceID,
		m.CreatedAt, m.UpdatedAt, m.DeletedAt,
	)
}

func documentModelFromEntity(d *entity.Document) *documentModel {
	return &documentModel{
		ID:           d.ID,
		OriginalName: d.OriginalName,
		MimeType:     d.MimeType,
		SizeBytes:    d.SizeBytes,
		StorageKey:   d.StorageKey,
		UploadedBy:   d.UploadedBy,
		Module:       d.Module,
		ReferenceID:  d.ReferenceID,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		DeletedAt:    d.DeletedAt,
	}
}

type postgresDocumentRepo struct {
	db *sqlx.DB
}

func NewDocumentRepo(db *sqlx.DB) DocumentRepository {
	return &postgresDocumentRepo{db: db}
}

func (r *postgresDocumentRepo) Create(ctx context.Context, doc *entity.Document) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	m := documentModelFromEntity(doc)

	query := `INSERT INTO documents (id, original_name, mime_type, size_bytes, storage_key,
	          uploaded_by, module, reference_id, created_at, updated_at)
	          VALUES (:id, :original_name, :mime_type, :size_bytes, :storage_key,
	          :uploaded_by, :module, :reference_id, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, m)
	return err
}

func (r *postgresDocumentRepo) FindByIDs(ctx context.Context, ids []string) (map[string]*entity.Document, error) {
	if len(ids) == 0 {
		return map[string]*entity.Document{}, nil
	}

	query := `SELECT id, original_name, mime_type, size_bytes, storage_key,
	          uploaded_by, module, reference_id, created_at, updated_at, deleted_at
	          FROM documents WHERE id IN (?) AND deleted_at IS NULL`

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	query = r.db.Rebind(query)

	var models []documentModel
	if err := r.db.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, fmt.Errorf("find documents by ids: %w", err)
	}

	result := make(map[string]*entity.Document, len(models))
	for i := range models {
		result[models[i].ID] = models[i].toEntity()
	}
	return result, nil
}

func (r *postgresDocumentRepo) FindByID(ctx context.Context, id string) (*entity.Document, error) {
	query := `SELECT id, original_name, mime_type, size_bytes, storage_key,
	          uploaded_by, module, reference_id, created_at, updated_at, deleted_at
	          FROM documents WHERE id = $1 AND deleted_at IS NULL`

	var m documentModel
	if err := r.db.GetContext(ctx, &m, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("find document by id: %w", err)
	}
	return m.toEntity(), nil
}

func (r *postgresDocumentRepo) ListByModule(ctx context.Context, module, referenceID string) ([]entity.Document, error) {
	query := `SELECT id, original_name, mime_type, size_bytes, storage_key,
	          uploaded_by, module, reference_id, created_at, updated_at, deleted_at
	          FROM documents WHERE module = $1 AND reference_id = $2 AND deleted_at IS NULL
	          ORDER BY created_at DESC`

	var models []documentModel
	if err := r.db.SelectContext(ctx, &models, query, module, referenceID); err != nil {
		return nil, fmt.Errorf("list documents by module: %w", err)
	}

	docs := make([]entity.Document, len(models))
	for i := range models {
		docs[i] = *models[i].toEntity()
	}
	return docs, nil
}

func (r *postgresDocumentRepo) CreateForModule(ctx context.Context, originalName, mimeType string, uploaderID *string, module, referenceID string, sizeBytes int64, id, storageKey string) error {
	if id == "" {
		id = uuid.New().String()
	}
	now := time.Now()
	query := `INSERT INTO documents (id, original_name, mime_type, size_bytes, storage_key,
	          uploaded_by, module, reference_id, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.ExecContext(ctx, query, id, originalName, mimeType, sizeBytes, storageKey,
		uploaderID, module, referenceID, now, now)
	return err
}

func (r *postgresDocumentRepo) DeleteSoft(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `UPDATE documents SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, now, id)
	return err
}
