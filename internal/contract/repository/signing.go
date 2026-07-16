package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/contract/entity"
)

type ContractSigningModel struct {
	ID              string    `db:"id"`
	ContractID      string    `db:"contract_id"`
	Party           string    `db:"party"`
	SignedBy        string    `db:"signed_by"`
	SignedByName    string    `db:"signed_by_name"`
	SignedByTitle   string    `db:"signed_by_title"`
	Place           string    `db:"place"`
	SignatureBase64 string    `db:"signature_base64"`
	SignedAt        time.Time `db:"signed_at"`
	CreatedAt       time.Time `db:"created_at"`
}

const (
	queryInsertContractSigning = `
		INSERT INTO contract_signings (id, contract_id, party, signed_by, signed_by_name, signed_by_title, place, signature_base64, signed_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`
	querySelectSigningsByContract = `
		SELECT id, contract_id, party, signed_by, signed_by_name, signed_by_title, place, signature_base64, signed_at, created_at
		FROM contract_signings WHERE contract_id=$1 ORDER BY signed_at ASC
	`
)

type PostgresSigningRepo struct {
	db txContext
}

func NewPostgresSigningRepo(db *sqlx.DB) *PostgresSigningRepo {
	return &PostgresSigningRepo{db: db}
}

func (r *PostgresSigningRepo) WithTx(tx *sqlx.Tx) SigningRepository {
	return &PostgresSigningRepo{db: tx}
}

func (r *PostgresSigningRepo) CreateContractSigning(ctx context.Context, e *entity.ContractSigning) error {
	_, err := r.db.ExecContext(ctx, queryInsertContractSigning,
		e.ID, e.ContractID, e.Party,
		e.SignedBy, e.SignedByName, e.SignedByTitle, e.Place,
		e.SignatureBase64, e.SignedAt, e.CreatedAt,
	)
	return err
}

func (r *PostgresSigningRepo) FindSigningsByContractID(ctx context.Context, contractID string) ([]*entity.ContractSigning, error) {
	rows, err := r.db.QueryxContext(ctx, querySelectSigningsByContract, contractID)
	if err != nil {
		return nil, fmt.Errorf("find signings: %w", err)
	}
	defer rows.Close()
	var list []*entity.ContractSigning
	for rows.Next() {
		var m ContractSigningModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("scan signing: %w", err)
		}
		list = append(list, modelToContractSigning(&m))
	}
	return list, rows.Err()
}

func (r *PostgresSigningRepo) FindSigningsByContractIDs(ctx context.Context, contractIDs []string) (map[string][]*entity.ContractSigning, error) {
	if len(contractIDs) == 0 {
		return map[string][]*entity.ContractSigning{}, nil
	}
	query, args, err := sqlx.In(`SELECT id, contract_id, party, signed_by, signed_by_name, signed_by_title, place, signature_base64, signed_at, created_at FROM contract_signings WHERE contract_id IN (?) ORDER BY signed_at ASC`, contractIDs)
	if err != nil {
		return nil, fmt.Errorf("build signings query: %w", err)
	}
	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find signings by ids: %w", err)
	}
	defer rows.Close()
	result := make(map[string][]*entity.ContractSigning)
	for rows.Next() {
		var m ContractSigningModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("scan signing: %w", err)
		}
		s := modelToContractSigning(&m)
		result[s.ContractID] = append(result[s.ContractID], s)
	}
	return result, rows.Err()
}

func modelToContractSigning(m *ContractSigningModel) *entity.ContractSigning {
	return entity.ReconstituteContractSigning(m.ID, m.Party, m.ContractID,
		m.SignedBy, m.SignedByName, m.SignedByTitle, m.Place,
		m.SignatureBase64, m.SignedAt, m.CreatedAt)
}
