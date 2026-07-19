package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
)

// ---------------------------------------------------------------------------
// BenefitTypeRepository
// ---------------------------------------------------------------------------

type PostgresBenefitTypeRepo struct {
	db *sqlx.DB
}

func NewPostgresBenefitTypeRepo(db *sqlx.DB) *PostgresBenefitTypeRepo {
	return &PostgresBenefitTypeRepo{db: db}
}

const qryInsertBenefitType = `
	INSERT INTO benefit_types (id, name, description, employer_contribution_type, employer_contribution_value,
		employee_contribution_type, employee_contribution_value, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

const qrySelectBenefitType = `
	SELECT id, name, description, employer_contribution_type, employer_contribution_value,
		employee_contribution_type, employee_contribution_value, is_active, created_at, updated_at
	FROM benefit_types
`

const qryUpdateBenefitType = `
	UPDATE benefit_types SET
		name = $1, description = $2, employer_contribution_type = $3,
		employer_contribution_value = $4, employee_contribution_type = $5,
		employee_contribution_value = $6, is_active = $7, updated_at = $8
	WHERE id = $9
`

const qryDeleteBenefitType = `DELETE FROM benefit_types WHERE id = $1`

func (r *PostgresBenefitTypeRepo) Create(ctx context.Context, bt *entity.BenefitType) error {
	_, err := r.db.ExecContext(ctx, qryInsertBenefitType,
		bt.ID, bt.Name, bt.Description,
		string(bt.EmployerContributionType), bt.EmployerContributionValue,
		string(bt.EmployeeContributionType), bt.EmployeeContributionValue,
		bt.IsActive, bt.CreatedAt, bt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert benefit type: %w", err)
	}
	return nil
}

func (r *PostgresBenefitTypeRepo) FindByID(ctx context.Context, id string) (*entity.BenefitType, error) {
	var m BenefitTypeModel
	err := r.db.QueryRowxContext(ctx, qrySelectBenefitType+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find benefit type by id: %w", err)
	}
	return benefitTypeModelToEntity(&m), nil
}

func (r *PostgresBenefitTypeRepo) FindByCode(ctx context.Context, code string) (*entity.BenefitType, error) {
	var m BenefitTypeModel
	err := r.db.QueryRowxContext(ctx, qrySelectBenefitType+` WHERE name = $1`, code).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find benefit type by code: %w", err)
	}
	return benefitTypeModelToEntity(&m), nil
}

func (r *PostgresBenefitTypeRepo) FindAll(ctx context.Context, filter BenefitTypeFilter) ([]*entity.BenefitType, int64, error) {
	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.IsActive != nil {
		where = fmt.Sprintf(" WHERE is_active = $%d", argIdx)
		args = append(args, *filter.IsActive)
		argIdx++
	}

	var total int64
	countQry := "SELECT COUNT(*) FROM benefit_types" + where
	if err := r.db.GetContext(ctx, &total, countQry, args...); err != nil {
		return nil, 0, fmt.Errorf("count benefit types: %w", err)
	}
	if total == 0 {
		return []*entity.BenefitType{}, 0, nil
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	orderQry := qrySelectBenefitType + where + " ORDER BY name ASC"
	orderQry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	var models []BenefitTypeModel
	if err := r.db.SelectContext(ctx, &models, orderQry, args...); err != nil {
		return nil, 0, fmt.Errorf("list benefit types: %w", err)
	}

	result := make([]*entity.BenefitType, len(models))
	for i := range models {
		result[i] = benefitTypeModelToEntity(&models[i])
	}
	return result, total, nil
}

func (r *PostgresBenefitTypeRepo) Update(ctx context.Context, bt *entity.BenefitType) error {
	res, err := r.db.ExecContext(ctx, qryUpdateBenefitType,
		bt.Name, bt.Description,
		string(bt.EmployerContributionType), bt.EmployerContributionValue,
		string(bt.EmployeeContributionType), bt.EmployeeContributionValue,
		bt.IsActive, bt.UpdatedAt, bt.ID,
	)
	if err != nil {
		return fmt.Errorf("update benefit type: %w", err)
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

func (r *PostgresBenefitTypeRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeleteBenefitType, id)
	if err != nil {
		return fmt.Errorf("delete benefit type: %w", err)
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

func benefitTypeModelToEntity(m *BenefitTypeModel) *entity.BenefitType {
	return entity.ReconstituteBenefitType(
		m.ID, m.Name, m.Description,
		m.EmployerContributionType, m.EmployerContributionValue,
		m.EmployeeContributionType, m.EmployeeContributionValue,
		m.IsActive, m.CreatedAt, m.UpdatedAt,
	)
}

func (r *PostgresBenefitTypeRepo) CountByTypeID(ctx context.Context, typeID string) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM employee_benefits WHERE benefit_type_id = $1`, typeID)
	if err != nil {
		return 0, fmt.Errorf("count employee benefits by type: %w", err)
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// EmployeeBenefitRepository
// ---------------------------------------------------------------------------

type PostgresEmployeeBenefitRepo struct {
	db *sqlx.DB
}

func NewPostgresEmployeeBenefitRepo(db *sqlx.DB) *PostgresEmployeeBenefitRepo {
	return &PostgresEmployeeBenefitRepo{db: db}
}

const qryInsertEmpBenefit = `
	INSERT INTO employee_benefits (id, employee_id, benefit_type_id, participant_number, effective_date, end_date, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

const qrySelectEmpBenefit = `
	SELECT id, employee_id, benefit_type_id, participant_number, effective_date, end_date, created_at, updated_at
	FROM employee_benefits
`

const qryUpdateEmpBenefit = `
	UPDATE employee_benefits SET
		employee_id = $1, benefit_type_id = $2, participant_number = $3, effective_date = $4,
		end_date = $5, updated_at = $6
	WHERE id = $7
`

const qryDeleteEmpBenefit = `DELETE FROM employee_benefits WHERE id = $1`

func (r *PostgresEmployeeBenefitRepo) Create(ctx context.Context, eb *entity.EmployeeBenefit) error {
	_, err := r.db.ExecContext(ctx, qryInsertEmpBenefit,
		eb.ID, eb.EmployeeID, eb.BenefitTypeID, eb.ParticipantNumber,
		eb.EffectiveDate, eb.EndDate, eb.CreatedAt, eb.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert employee benefit: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeBenefitRepo) FindByID(ctx context.Context, id string) (*entity.EmployeeBenefit, error) {
	var m EmployeeBenefitModel
	err := r.db.QueryRowxContext(ctx, qrySelectEmpBenefit+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find employee benefit by id: %w", err)
	}
	return empBenefitModelToEntity(&m), nil
}

func (r *PostgresEmployeeBenefitRepo) FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeBenefit, error) {
	var models []EmployeeBenefitModel
	err := r.db.SelectContext(ctx, &models, qrySelectEmpBenefit+` WHERE employee_id = $1 ORDER BY effective_date DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee benefits by employee: %w", err)
	}
	result := make([]*entity.EmployeeBenefit, len(models))
	for i := range models {
		result[i] = empBenefitModelToEntity(&models[i])
	}
	return result, nil
}

func (r *PostgresEmployeeBenefitRepo) FindAll(ctx context.Context, filter EmpBenefitFilter) ([]*entity.EmployeeBenefit, int64, error) {
	where := ""
	args := []interface{}{}
	argIdx := 1

	if filter.EmployeeID != "" {
		where = fmt.Sprintf(" WHERE employee_id = $%d", argIdx)
		args = append(args, filter.EmployeeID)
		argIdx++
	}

	var total int64
	countQry := "SELECT COUNT(*) FROM employee_benefits" + where
	if err := r.db.GetContext(ctx, &total, countQry, args...); err != nil {
		return nil, 0, fmt.Errorf("count employee benefits: %w", err)
	}
	if total == 0 {
		return []*entity.EmployeeBenefit{}, 0, nil
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	orderQry := qrySelectEmpBenefit + where + " ORDER BY effective_date DESC"
	orderQry += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	var models []EmployeeBenefitModel
	if err := r.db.SelectContext(ctx, &models, orderQry, args...); err != nil {
		return nil, 0, fmt.Errorf("list employee benefits: %w", err)
	}

	result := make([]*entity.EmployeeBenefit, len(models))
	for i := range models {
		result[i] = empBenefitModelToEntity(&models[i])
	}
	return result, total, nil
}

func (r *PostgresEmployeeBenefitRepo) Update(ctx context.Context, eb *entity.EmployeeBenefit) error {
	res, err := r.db.ExecContext(ctx, qryUpdateEmpBenefit,
		eb.EmployeeID, eb.BenefitTypeID, eb.ParticipantNumber, eb.EffectiveDate, eb.EndDate, eb.UpdatedAt, eb.ID,
	)
	if err != nil {
		return fmt.Errorf("update employee benefit: %w", err)
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

func (r *PostgresEmployeeBenefitRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeleteEmpBenefit, id)
	if err != nil {
		return fmt.Errorf("delete employee benefit: %w", err)
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

func empBenefitModelToEntity(m *EmployeeBenefitModel) *entity.EmployeeBenefit {
	return entity.ReconstituteEmployeeBenefit(
		m.ID, m.EmployeeID, m.BenefitTypeID, m.ParticipantNumber,
		m.EffectiveDate, m.EndDate, m.CreatedAt, m.UpdatedAt,
	)
}
