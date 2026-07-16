package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
)

// ---------------------------------------------------------------------------
// DeductionTypeRepository
// ---------------------------------------------------------------------------

type PostgresDeductionTypeRepo struct {
	db *sqlx.DB
}

func NewPostgresDeductionTypeRepo(db *sqlx.DB) *PostgresDeductionTypeRepo {
	return &PostgresDeductionTypeRepo{db: db}
}

const qryInsertDeductionType = `
	INSERT INTO deduction_types (id, name, slug, description, deduction_type, default_value, is_active, is_mandatory, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

const qrySelectDeductionType = `
	SELECT id, name, slug, description, deduction_type, default_value, is_active, is_mandatory, created_at, updated_at
	FROM deduction_types
`

const qryUpdateDeductionType = `
	UPDATE deduction_types SET
		name = $1, slug = $2, description = $3, deduction_type = $4,
		default_value = $5, is_active = $6, is_mandatory = $7, updated_at = $8
	WHERE id = $9
`

const qryDeleteDeductionType = `DELETE FROM deduction_types WHERE id = $1`

func (r *PostgresDeductionTypeRepo) Create(ctx context.Context, dt *entity.DeductionType) error {
	_, err := r.db.ExecContext(ctx, qryInsertDeductionType,
		dt.ID, dt.Name, dt.Slug, dt.Description, string(dt.DeductionType),
		dt.DefaultValue, dt.IsActive, dt.IsMandatory, dt.CreatedAt, dt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert deduction type: %w", err)
	}
	return nil
}

func (r *PostgresDeductionTypeRepo) FindByID(ctx context.Context, id string) (*entity.DeductionType, error) {
	var m DeductionTypeModel
	err := r.db.QueryRowxContext(ctx, qrySelectDeductionType+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find deduction type by id: %w", err)
	}
	return deductionTypeModelToEntity(&m), nil
}

func (r *PostgresDeductionTypeRepo) FindByCode(ctx context.Context, code string) (*entity.DeductionType, error) {
	var m DeductionTypeModel
	err := r.db.QueryRowxContext(ctx, qrySelectDeductionType+` WHERE name = $1`, code).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find deduction type by code: %w", err)
	}
	return deductionTypeModelToEntity(&m), nil
}

func (r *PostgresDeductionTypeRepo) FindBySlug(ctx context.Context, slug string) (*entity.DeductionType, error) {
	var m DeductionTypeModel
	err := r.db.QueryRowxContext(ctx, qrySelectDeductionType+` WHERE slug = $1`, slug).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find deduction type by slug: %w", err)
	}
	return deductionTypeModelToEntity(&m), nil
}

func (r *PostgresDeductionTypeRepo) FindAll(ctx context.Context, filter DeductionTypeFilter) ([]*entity.DeductionType, int64, error) {
	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.IsActive != nil {
		where = addWhere(where, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}
	if filter.IsMandatory != nil {
		where = addWhere(where, fmt.Sprintf("is_mandatory = $%d", argIdx))
		args = append(args, *filter.IsMandatory)
		argIdx++
	}

	var total int64
	countQry := "SELECT COUNT(*) FROM deduction_types" + where
	if err := r.db.GetContext(ctx, &total, countQry, args...); err != nil {
		return nil, 0, fmt.Errorf("count deduction types: %w", err)
	}
	if total == 0 {
		return []*entity.DeductionType{}, 0, nil
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	orderQry := qrySelectDeductionType + where + " ORDER BY name ASC"
	orderQry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	var models []DeductionTypeModel
	if err := r.db.SelectContext(ctx, &models, orderQry, args...); err != nil {
		return nil, 0, fmt.Errorf("list deduction types: %w", err)
	}

	result := make([]*entity.DeductionType, len(models))
	for i := range models {
		result[i] = deductionTypeModelToEntity(&models[i])
	}
	return result, total, nil
}

func (r *PostgresDeductionTypeRepo) Update(ctx context.Context, dt *entity.DeductionType) error {
	res, err := r.db.ExecContext(ctx, qryUpdateDeductionType,
		dt.Name, dt.Slug, dt.Description, string(dt.DeductionType),
		dt.DefaultValue, dt.IsActive, dt.IsMandatory, dt.UpdatedAt, dt.ID,
	)
	if err != nil {
		return fmt.Errorf("update deduction type: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("deduction type not found")
	}
	return nil
}

func (r *PostgresDeductionTypeRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeleteDeductionType, id)
	if err != nil {
		return fmt.Errorf("delete deduction type: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("deduction type not found")
	}
	return nil
}

func deductionTypeModelToEntity(m *DeductionTypeModel) *entity.DeductionType {
	return entity.ReconstituteDeductionType(
		m.ID, m.Name, m.Slug, m.Description,
		m.DeductionType, m.DefaultValue, m.IsActive, m.IsMandatory,
		m.CreatedAt, m.UpdatedAt,
	)
}

func (r *PostgresDeductionTypeRepo) CountByTypeID(ctx context.Context, typeID string) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM employee_deductions WHERE deduction_type_id = $1`, typeID)
	if err != nil {
		return 0, fmt.Errorf("count employee deductions by type: %w", err)
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// EmployeeDeductionRepository
// ---------------------------------------------------------------------------

type PostgresEmployeeDeductionRepo struct {
	db *sqlx.DB
}

func NewPostgresEmployeeDeductionRepo(db *sqlx.DB) *PostgresEmployeeDeductionRepo {
	return &PostgresEmployeeDeductionRepo{db: db}
}

const qryInsertEmpDeduction = `
	INSERT INTO employee_deductions (id, employee_id, deduction_type_id, value, effective_date, end_date, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

const qrySelectEmpDeduction = `
	SELECT id, employee_id, deduction_type_id, value, effective_date, end_date, created_at, updated_at
	FROM employee_deductions
`

const qryUpdateEmpDeduction = `
	UPDATE employee_deductions SET
		employee_id = $1, deduction_type_id = $2, value = $3,
		effective_date = $4, end_date = $5, updated_at = $6
	WHERE id = $7
`

const qryDeleteEmpDeduction = `DELETE FROM employee_deductions WHERE id = $1`

func (r *PostgresEmployeeDeductionRepo) Create(ctx context.Context, ed *entity.EmployeeDeduction) error {
	_, err := r.db.ExecContext(ctx, qryInsertEmpDeduction,
		ed.ID, ed.EmployeeID, ed.DeductionTypeID, ed.Value,
		ed.EffectiveDate, ed.EndDate, ed.CreatedAt, ed.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert employee deduction: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeDeductionRepo) FindByID(ctx context.Context, id string) (*entity.EmployeeDeduction, error) {
	var m EmployeeDeductionModel
	err := r.db.QueryRowxContext(ctx, qrySelectEmpDeduction+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find employee deduction by id: %w", err)
	}
	return empDeductionModelToEntity(&m), nil
}

func (r *PostgresEmployeeDeductionRepo) FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeDeduction, error) {
	var models []EmployeeDeductionModel
	err := r.db.SelectContext(ctx, &models, qrySelectEmpDeduction+` WHERE employee_id = $1 ORDER BY effective_date DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee deductions by employee: %w", err)
	}
	result := make([]*entity.EmployeeDeduction, len(models))
	for i := range models {
		result[i] = empDeductionModelToEntity(&models[i])
	}
	return result, nil
}

func (r *PostgresEmployeeDeductionRepo) FindAll(ctx context.Context, filter EmpDeductionFilter) ([]*entity.EmployeeDeduction, int64, error) {
	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.EmployeeID != "" {
		where = fmt.Sprintf(" WHERE employee_id = $%d", argIdx)
		args = append(args, filter.EmployeeID)
		argIdx++
	}

	var total int64
	countQry := "SELECT COUNT(*) FROM employee_deductions" + where
	if err := r.db.GetContext(ctx, &total, countQry, args...); err != nil {
		return nil, 0, fmt.Errorf("count employee deductions: %w", err)
	}
	if total == 0 {
		return []*entity.EmployeeDeduction{}, 0, nil
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	orderQry := qrySelectEmpDeduction + where + " ORDER BY effective_date DESC"
	orderQry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	var models []EmployeeDeductionModel
	if err := r.db.SelectContext(ctx, &models, orderQry, args...); err != nil {
		return nil, 0, fmt.Errorf("list employee deductions: %w", err)
	}

	result := make([]*entity.EmployeeDeduction, len(models))
	for i := range models {
		result[i] = empDeductionModelToEntity(&models[i])
	}
	return result, total, nil
}

func (r *PostgresEmployeeDeductionRepo) Update(ctx context.Context, ed *entity.EmployeeDeduction) error {
	res, err := r.db.ExecContext(ctx, qryUpdateEmpDeduction,
		ed.EmployeeID, ed.DeductionTypeID, ed.Value,
		ed.EffectiveDate, ed.EndDate, ed.UpdatedAt, ed.ID,
	)
	if err != nil {
		return fmt.Errorf("update employee deduction: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("employee deduction not found")
	}
	return nil
}

func (r *PostgresEmployeeDeductionRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeleteEmpDeduction, id)
	if err != nil {
		return fmt.Errorf("delete employee deduction: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("employee deduction not found")
	}
	return nil
}

func empDeductionModelToEntity(m *EmployeeDeductionModel) *entity.EmployeeDeduction {
	return entity.ReconstituteEmployeeDeduction(
		m.ID, m.EmployeeID, m.DeductionTypeID, m.Value,
		m.EffectiveDate, m.EndDate, m.CreatedAt, m.UpdatedAt,
	)
}

// addWhere is a small helper to build WHERE clauses incrementally.
func addWhere(current string, condition string) string {
	if current == "" {
		return " WHERE " + condition
	}
	return current + " AND " + condition
}
