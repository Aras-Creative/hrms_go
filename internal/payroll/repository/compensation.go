package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
)

// ---------------------------------------------------------------------------
// CompensationItemRepository
// ---------------------------------------------------------------------------

type PostgresCompensationItemRepo struct {
	db *sqlx.DB
}

func NewPostgresCompensationItemRepo(db *sqlx.DB) *PostgresCompensationItemRepo {
	return &PostgresCompensationItemRepo{db: db}
}

const qryInsertCompItem = `
	INSERT INTO compensation_items (id, name, item_type, description, is_active, is_taxable, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

const qrySelectCompItem = `
	SELECT id, name, item_type, description, is_active, is_taxable, created_at, updated_at
	FROM compensation_items
`

const qryUpdateCompItem = `
	UPDATE compensation_items SET
		name = $1, item_type = $2, description = $3,
		is_active = $4, is_taxable = $5, updated_at = $6
	WHERE id = $7
`

const qryDeleteCompItem = `DELETE FROM compensation_items WHERE id = $1`

func (r *PostgresCompensationItemRepo) Create(ctx context.Context, ci *entity.CompensationItem) error {
	_, err := r.db.ExecContext(ctx, qryInsertCompItem,
		ci.ID, ci.Name, string(ci.ItemType), ci.Description, ci.IsActive, ci.IsTaxable, ci.CreatedAt, ci.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert compensation item: %w", err)
	}
	return nil
}

func (r *PostgresCompensationItemRepo) FindByID(ctx context.Context, id string) (*entity.CompensationItem, error) {
	var m CompensationItemModel
	err := r.db.QueryRowxContext(ctx, qrySelectCompItem+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find compensation item by id: %w", err)
	}
	return compItemModelToEntity(&m), nil
}

func (r *PostgresCompensationItemRepo) FindByCode(ctx context.Context, code string) (*entity.CompensationItem, error) {
	var m CompensationItemModel
	err := r.db.QueryRowxContext(ctx, qrySelectCompItem+` WHERE name = $1`, code).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find compensation item by code: %w", err)
	}
	return compItemModelToEntity(&m), nil
}

func (r *PostgresCompensationItemRepo) FindAll(ctx context.Context, filter CompItemFilter) ([]*entity.CompensationItem, int64, error) {
	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.IsActive != nil {
		where = fmt.Sprintf(" WHERE is_active = $%d", argIdx)
		args = append(args, *filter.IsActive)
		argIdx++
	}

	var total int64
	countQry := "SELECT COUNT(*) FROM compensation_items" + where
	if err := r.db.GetContext(ctx, &total, countQry, args...); err != nil {
		return nil, 0, fmt.Errorf("count compensation items: %w", err)
	}
	if total == 0 {
		return []*entity.CompensationItem{}, 0, nil
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	orderQry := qrySelectCompItem + where + " ORDER BY name ASC"
	orderQry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	var models []CompensationItemModel
	if err := r.db.SelectContext(ctx, &models, orderQry, args...); err != nil {
		return nil, 0, fmt.Errorf("list compensation items: %w", err)
	}

	result := make([]*entity.CompensationItem, len(models))
	for i := range models {
		result[i] = compItemModelToEntity(&models[i])
	}
	return result, total, nil
}

func (r *PostgresCompensationItemRepo) Update(ctx context.Context, ci *entity.CompensationItem) error {
	res, err := r.db.ExecContext(ctx, qryUpdateCompItem,
		ci.Name, string(ci.ItemType), ci.Description, ci.IsActive, ci.IsTaxable, ci.UpdatedAt, ci.ID,
	)
	if err != nil {
		return fmt.Errorf("update compensation item: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return nil
	}
	return nil
}

func (r *PostgresCompensationItemRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeleteCompItem, id)
	if err != nil {
		return fmt.Errorf("delete compensation item: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return nil
	}
	return nil
}

func compItemModelToEntity(m *CompensationItemModel) *entity.CompensationItem {
	return entity.ReconstituteCompensationItem(
		m.ID, m.Name, m.ItemType, m.Description, m.IsActive, m.IsTaxable,
		m.CreatedAt, m.UpdatedAt,
	)
}

func (r *PostgresCompensationItemRepo) CountByItemID(ctx context.Context, itemID string) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM employee_compensations WHERE compensation_item_id = $1`, itemID)
	if err != nil {
		return 0, fmt.Errorf("count employee compensations by item: %w", err)
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// EmployeeCompensationRepository
// ---------------------------------------------------------------------------

type PostgresEmployeeCompensationRepo struct {
	db *sqlx.DB
}

func NewPostgresEmployeeCompensationRepo(db *sqlx.DB) *PostgresEmployeeCompensationRepo {
	return &PostgresEmployeeCompensationRepo{db: db}
}

const qryInsertEmpComp = `
	INSERT INTO employee_compensations (id, employee_id, compensation_item_id, amount, frequency, effective_date, end_date, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

const qrySelectEmpComp = `
	SELECT id, employee_id, compensation_item_id, amount, frequency, effective_date, end_date, created_at, updated_at
	FROM employee_compensations
`

const qryUpdateEmpComp = `
	UPDATE employee_compensations SET
		employee_id = $1, compensation_item_id = $2, amount = $3, frequency = $4,
		effective_date = $5, end_date = $6, updated_at = $7
	WHERE id = $8
`

const qryDeleteEmpComp = `DELETE FROM employee_compensations WHERE id = $1`

func (r *PostgresEmployeeCompensationRepo) Create(ctx context.Context, ec *entity.EmployeeCompensation) error {
	_, err := r.db.ExecContext(ctx, qryInsertEmpComp,
		ec.ID, ec.EmployeeID, ec.CompensationItemID, ec.Amount.Cents(), string(ec.Frequency),
		ec.EffectiveDate, ec.EndDate, ec.CreatedAt, ec.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert employee compensation: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeCompensationRepo) FindByID(ctx context.Context, id string) (*entity.EmployeeCompensation, error) {
	var m EmployeeCompensationModel
	err := r.db.QueryRowxContext(ctx, qrySelectEmpComp+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find employee compensation by id: %w", err)
	}
	return empCompModelToEntity(&m), nil
}

func (r *PostgresEmployeeCompensationRepo) FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeCompensation, error) {
	var models []EmployeeCompensationModel
	err := r.db.SelectContext(ctx, &models, qrySelectEmpComp+` WHERE employee_id = $1 ORDER BY effective_date DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee compensations by employee: %w", err)
	}
	result := make([]*entity.EmployeeCompensation, len(models))
	for i := range models {
		result[i] = empCompModelToEntity(&models[i])
	}
	return result, nil
}

func (r *PostgresEmployeeCompensationRepo) FindAll(ctx context.Context, filter EmpCompFilter) ([]*entity.EmployeeCompensation, int64, error) {
	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.EmployeeID != "" {
		where = fmt.Sprintf(" WHERE employee_id = $%d", argIdx)
		args = append(args, filter.EmployeeID)
		argIdx++
	}

	var total int64
	countQry := "SELECT COUNT(*) FROM employee_compensations" + where
	if err := r.db.GetContext(ctx, &total, countQry, args...); err != nil {
		return nil, 0, fmt.Errorf("count employee compensations: %w", err)
	}
	if total == 0 {
		return []*entity.EmployeeCompensation{}, 0, nil
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	orderQry := qrySelectEmpComp + where + " ORDER BY effective_date DESC"
	orderQry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	var models []EmployeeCompensationModel
	if err := r.db.SelectContext(ctx, &models, orderQry, args...); err != nil {
		return nil, 0, fmt.Errorf("list employee compensations: %w", err)
	}

	result := make([]*entity.EmployeeCompensation, len(models))
	for i := range models {
		result[i] = empCompModelToEntity(&models[i])
	}
	return result, total, nil
}

func (r *PostgresEmployeeCompensationRepo) Update(ctx context.Context, ec *entity.EmployeeCompensation) error {
	res, err := r.db.ExecContext(ctx, qryUpdateEmpComp,
		ec.EmployeeID, ec.CompensationItemID, ec.Amount.Cents(), string(ec.Frequency),
		ec.EffectiveDate, ec.EndDate, ec.UpdatedAt, ec.ID,
	)
	if err != nil {
		return fmt.Errorf("update employee compensation: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return nil
	}
	return nil
}

func (r *PostgresEmployeeCompensationRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeleteEmpComp, id)
	if err != nil {
		return fmt.Errorf("delete employee compensation: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return nil
	}
	return nil
}

func empCompModelToEntity(m *EmployeeCompensationModel) *entity.EmployeeCompensation {
	return entity.ReconstituteEmployeeCompensation(
		m.ID, m.EmployeeID, m.CompensationItemID, m.Amount, m.Frequency,
		m.EffectiveDate, m.EndDate, m.CreatedAt, m.UpdatedAt,
	)
}
