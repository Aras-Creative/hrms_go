package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/contract/entity"
)

type ContractDocumentModel struct {
	ID          string    `db:"id"`
	ContractID  string    `db:"contract_id"`
	DocumentID  string    `db:"document_id"`
	ContentHash string    `db:"content_hash"`
	CreatedAt   time.Time `db:"created_at"`
}

const (
	queryInsertContractDocument = `
		INSERT INTO contract_documents (id, contract_id, document_id, content_hash, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (contract_id) DO UPDATE SET document_id=EXCLUDED.document_id, content_hash=EXCLUDED.content_hash, created_at=EXCLUDED.created_at
	`
	querySelectContractDocumentByContract = `
		SELECT id, contract_id, document_id, content_hash, created_at
		FROM contract_documents WHERE contract_id=$1
	`
)

type PostgresDocumentRepo struct {
	db *sqlx.DB
}

func NewPostgresDocumentRepo(db *sqlx.DB) *PostgresDocumentRepo {
	return &PostgresDocumentRepo{db: db}
}

func (r *PostgresDocumentRepo) CreateContractDocument(ctx context.Context, e *entity.ContractDocument) error {
	_, err := r.db.ExecContext(ctx, queryInsertContractDocument,
		e.ID, e.ContractID, e.DocumentID, e.ContentHash, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create contract document: %w", err)
	}
	return nil
}

func (r *PostgresDocumentRepo) FindContractDocumentByContractID(ctx context.Context, contractID string) (*entity.ContractDocument, error) {
	var m ContractDocumentModel
	err := r.db.QueryRowxContext(ctx, querySelectContractDocumentByContract, contractID).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find contract document: %w", err)
	}
	return entity.ReconstituteContractDocument(m.ID, m.ContractID, m.DocumentID, m.ContentHash, m.CreatedAt), nil
}
