package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
)

type ContractTemplateModel struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	ContractType string    `db:"contract_type"`
	Description  string    `db:"description"`
	IsActive     bool      `db:"is_active"`
	Data         []byte    `db:"data"`
	Templates    []byte    `db:"templates"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

const (
	queryInsertTemplate = `
		INSERT INTO contract_templates (id, name, contract_type, description, is_active, data, templates, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	querySelectTemplate = `
		SELECT id, name, contract_type, description, is_active, data, templates, created_at, updated_at
		FROM contract_templates
	`
	queryUpdateTemplate = `
		UPDATE contract_templates SET name=$1, contract_type=$2, description=$3, is_active=$4, data=$5, templates=$6, updated_at=$7 WHERE id=$8 AND updated_at=$9
	`
	queryDeleteTemplate = `DELETE FROM contract_templates WHERE id=$1`
)

var (
	queryTemplateByID    = querySelectTemplate + ` WHERE id=$1`
	queryTemplateByName  = querySelectTemplate + ` WHERE name=$1`
	queryTemplateAll     = querySelectTemplate
	queryTemplateCount   = `SELECT COUNT(*) FROM contract_templates`
	queryCountContractsByTemplate = `SELECT COUNT(*) FROM contracts WHERE template_id=$1`
)

func modelToTemplate(m *ContractTemplateModel) (*entity.ContractTemplate, error) {
	var data entity.ContractTemplateData
	if len(m.Data) > 0 {
		if err := json.Unmarshal(m.Data, &data); err != nil {
			return nil, fmt.Errorf("unmarshal template data: %w", err)
		}
	}
	var templates entity.ContractTemplatePartials
	if len(m.Templates) > 0 {
		if err := json.Unmarshal(m.Templates, &templates); err != nil {
			return nil, fmt.Errorf("unmarshal template partials: %w", err)
		}
	}
	return entity.ReconstituteContractTemplate(m.ID, m.Name, entity.ContractType(m.ContractType), m.Description, m.IsActive, data, templates, m.CreatedAt, m.UpdatedAt), nil
}

type PostgresTemplateRepo struct {
	db *sqlx.DB
}

func NewPostgresTemplateRepo(db *sqlx.DB) *PostgresTemplateRepo {
	return &PostgresTemplateRepo{db: db}
}

func (r *PostgresTemplateRepo) Create(ctx context.Context, e *entity.ContractTemplate) error {
	dataJSON, err := json.Marshal(e.Data)
	if err != nil {
		return fmt.Errorf("marshal template data: %w", err)
	}
	templatesJSON, err := json.Marshal(e.Templates)
	if err != nil {
		return fmt.Errorf("marshal template partials: %w", err)
	}
	_, err = r.db.ExecContext(ctx, queryInsertTemplate, e.ID, e.Name, string(e.ContractType), e.Description, e.IsActive, dataJSON, templatesJSON, e.CreatedAt, e.UpdatedAt)
	return err
}

func (r *PostgresTemplateRepo) FindByID(ctx context.Context, id string) (*entity.ContractTemplate, error) {
	var m ContractTemplateModel
	err := r.db.QueryRowxContext(ctx, queryTemplateByID, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find template: %w", err)
	}
	return modelToTemplate(&m)
}

func (r *PostgresTemplateRepo) FindByName(ctx context.Context, name string) (*entity.ContractTemplate, error) {
	var m ContractTemplateModel
	err := r.db.QueryRowxContext(ctx, queryTemplateByName, name).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find template by name: %w", err)
	}
	return modelToTemplate(&m)
}

func (r *PostgresTemplateRepo) FindAll(ctx context.Context, filter models.ListTemplateInput) ([]*entity.ContractTemplate, int64, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1
	if filter.SearchName != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", argIdx)
		args = append(args, "%"+filter.SearchName+"%")
		argIdx++
	}
	if filter.ContractType != "" {
		where += fmt.Sprintf(" AND contract_type=$%d", argIdx)
		args = append(args, filter.ContractType)
		argIdx++
	}
	if filter.IsActive != nil {
		where += fmt.Sprintf(" AND is_active=$%d", argIdx)
		args = append(args, *filter.IsActive)
		argIdx++
	}

	var total int64
	countQuery := queryTemplateCount + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count templates: %w", err)
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := queryTemplateAll + where + fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var list []*entity.ContractTemplate
	for rows.Next() {
		var m ContractTemplateModel
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("scan template: %w", err)
		}
		t, err := modelToTemplate(&m)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, t)
	}
	return list, total, rows.Err()
}

func (r *PostgresTemplateRepo) Update(ctx context.Context, e *entity.ContractTemplate, expectedUpdatedAt time.Time) error {
	dataJSON, err := json.Marshal(e.Data)
	if err != nil {
		return fmt.Errorf("marshal template data: %w", err)
	}
	templatesJSON, err := json.Marshal(e.Templates)
	if err != nil {
		return fmt.Errorf("marshal template partials: %w", err)
	}
	result, err := r.db.ExecContext(ctx, queryUpdateTemplate, e.Name, string(e.ContractType), e.Description, e.IsActive, dataJSON, templatesJSON, e.UpdatedAt, e.ID, expectedUpdatedAt)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update template rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("template not found or modified by another user")
	}
	return nil
}

func (r *PostgresTemplateRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, queryDeleteTemplate, id)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete template rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("template not found")
	}
	return nil
}

func (r *PostgresTemplateRepo) CountByTemplateID(ctx context.Context, templateID string) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, queryCountContractsByTemplate, templateID)
	if err != nil {
		return 0, fmt.Errorf("count contracts by template: %w", err)
	}
	return count, nil
}
