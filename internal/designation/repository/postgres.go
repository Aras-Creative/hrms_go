package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/designation/entity"
	"hrms/internal/designation/models"
)

const (
	queryInsertDesignation = `
		INSERT INTO designations (id, code, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	querySelectDesignation = `
		SELECT id, code, name, created_at, updated_at
		FROM designations
	`

	queryUpdateDesignation = `
		UPDATE designations
		SET name = $1, updated_at = $2
		WHERE id = $3
	`

	queryDeleteDesignation = `DELETE FROM designations WHERE id = $1`

	querySelectMembers = `
		SELECT id, full_name, employee_number, profile_photo_id, designation_id
		FROM employees
		WHERE designation_id IS NOT NULL
		ORDER BY full_name ASC
	`
)

var (
	queryDesignationByID   = querySelectDesignation + ` WHERE id = $1`
	queryDesignationByCode = querySelectDesignation + ` WHERE code = $1`
	queryDesignationAll    = querySelectDesignation + ` ORDER BY name ASC`
)

type PostgresDesignationRepo struct {
	db *sqlx.DB
}

func NewPostgresDesignationRepo(db *sqlx.DB) *PostgresDesignationRepo {
	return &PostgresDesignationRepo{db: db}
}

func (r *PostgresDesignationRepo) Create(ctx context.Context, d *entity.Designation) error {
	_, err := r.db.ExecContext(ctx, queryInsertDesignation,
		d.ID, d.Code, d.Name, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create designation: %w", err)
	}
	return nil
}

func (r *PostgresDesignationRepo) FindByID(ctx context.Context, id string) (*entity.Designation, error) {
	var d entity.Designation
	err := r.db.QueryRowxContext(ctx, queryDesignationByID, id).Scan(
		&d.ID, &d.Code, &d.Name, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find designation: %w", err)
	}
	return &d, nil
}

func (r *PostgresDesignationRepo) FindByIDs(ctx context.Context, ids []string) ([]*entity.Designation, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(querySelectDesignation+` WHERE id IN (?)`, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find designations by ids: %w", err)
	}
	defer rows.Close()

	var list []*entity.Designation
	for rows.Next() {
		var d entity.Designation
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan designation: %w", err)
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *PostgresDesignationRepo) FindByCode(ctx context.Context, code string) (*entity.Designation, error) {
	var d entity.Designation
	err := r.db.QueryRowxContext(ctx, queryDesignationByCode, code).Scan(
		&d.ID, &d.Code, &d.Name, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find designation by code: %w", err)
	}
	return &d, nil
}

func (r *PostgresDesignationRepo) FindAll(ctx context.Context) ([]models.DesignationReadModel, error) {
	rows, err := r.db.QueryxContext(ctx, queryDesignationAll)
	if err != nil {
		return nil, fmt.Errorf("failed to list designations: %w", err)
	}
	defer rows.Close()

	var list []models.DesignationReadModel
	for rows.Next() {
		var d models.DesignationReadModel
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan designation: %w", err)
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	members, err := r.db.QueryxContext(ctx, querySelectMembers)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer members.Close()

	grouped := make(map[string][]models.MemberRow)
	for members.Next() {
		var m models.MemberRow
		if err := members.Scan(&m.ID, &m.FullName, &m.EmployeeNumber, &m.ProfilePhotoID, &m.DesignationID); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		grouped[m.DesignationID] = append(grouped[m.DesignationID], m)
	}
	if err := members.Err(); err != nil {
		return nil, err
	}

	for i := range list {
		list[i].Members = grouped[list[i].ID]
		list[i].MemberCount = len(list[i].Members)
	}
	return list, nil
}

func (r *PostgresDesignationRepo) Update(ctx context.Context, d *entity.Designation) error {
	result, err := r.db.ExecContext(ctx, queryUpdateDesignation, d.Name, d.UpdatedAt, d.ID)
	if err != nil {
		return fmt.Errorf("failed to update designation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("designation with id %s not found", d.ID)
	}
	return nil
}

func (r *PostgresDesignationRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, queryDeleteDesignation, id)
	if err != nil {
		return fmt.Errorf("failed to delete designation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("designation with id %s not found", id)
	}
	return nil
}
