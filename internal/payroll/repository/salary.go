package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
)

// ---------------------------------------------------------------------------
// EmployeeBaseSalaryRepository
// ---------------------------------------------------------------------------

type PostgresEmployeeBaseSalaryRepo struct {
	db *sqlx.DB
}

func NewPostgresEmployeeBaseSalaryRepo(db *sqlx.DB) *PostgresEmployeeBaseSalaryRepo {
	return &PostgresEmployeeBaseSalaryRepo{db: db}
}

const qryInsertBaseSalary = `
	INSERT INTO employee_base_salaries (id, employee_id, amount, currency, effective_date, end_date, notes, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

const qrySelectBaseSalary = `
	SELECT id, employee_id, amount, currency, effective_date, end_date, notes, created_at, updated_at
	FROM employee_base_salaries
`

const qryUpdateBaseSalary = `
	UPDATE employee_base_salaries SET
		employee_id = $1, amount = $2, currency = $3, effective_date = $4,
		end_date = $5, notes = $6, updated_at = $7
	WHERE id = $8
`

const qryDeleteBaseSalary = `DELETE FROM employee_base_salaries WHERE id = $1`

func (r *PostgresEmployeeBaseSalaryRepo) Create(ctx context.Context, s *entity.EmployeeBaseSalary) error {
	_, err := r.db.ExecContext(ctx, qryInsertBaseSalary,
		s.ID, s.EmployeeID, s.Amount.Cents(), s.Currency.String(),
		s.EffectiveDate, s.EndDate, s.Notes, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert base salary: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeBaseSalaryRepo) FindByID(ctx context.Context, id string) (*entity.EmployeeBaseSalary, error) {
	var m EmployeeBaseSalaryModel
	err := r.db.QueryRowxContext(ctx, qrySelectBaseSalary+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find base salary by id: %w", err)
	}
	return baseSalaryModelToEntity(&m), nil
}

func (r *PostgresEmployeeBaseSalaryRepo) FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeBaseSalary, error) {
	var models []EmployeeBaseSalaryModel
	err := r.db.SelectContext(ctx, &models, qrySelectBaseSalary+` WHERE employee_id = $1 ORDER BY effective_date DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("find base salaries by employee: %w", err)
	}
	result := make([]*entity.EmployeeBaseSalary, len(models))
	for i := range models {
		result[i] = baseSalaryModelToEntity(&models[i])
	}
	return result, nil
}

func (r *PostgresEmployeeBaseSalaryRepo) FindCurrentByEmployeeID(ctx context.Context, employeeID string) (*entity.EmployeeBaseSalary, error) {
	var m EmployeeBaseSalaryModel
	err := r.db.QueryRowxContext(ctx, qrySelectBaseSalary+` WHERE employee_id = $1 AND end_date IS NULL ORDER BY effective_date DESC LIMIT 1`, employeeID).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find current base salary: %w", err)
	}
	return baseSalaryModelToEntity(&m), nil
}

func (r *PostgresEmployeeBaseSalaryRepo) FindCurrentByEmployeeIDs(ctx context.Context, employeeIDs []string) (map[string]*entity.EmployeeBaseSalary, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(qrySelectBaseSalary+` WHERE employee_id IN (?) AND end_date IS NULL`, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find current base salaries: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*entity.EmployeeBaseSalary)
	for rows.Next() {
		var m EmployeeBaseSalaryModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("scan base salary: %w", err)
		}
		if existing, ok := result[m.EmployeeID]; !ok || m.EffectiveDate.After(existing.EffectiveDate) {
			result[m.EmployeeID] = baseSalaryModelToEntity(&m)
		}
	}
	return result, rows.Err()
}

func (r *PostgresEmployeeBaseSalaryRepo) Update(ctx context.Context, s *entity.EmployeeBaseSalary) error {
	res, err := r.db.ExecContext(ctx, qryUpdateBaseSalary,
		s.EmployeeID, s.Amount.Cents(), s.Currency.String(),
		s.EffectiveDate, s.EndDate, s.Notes, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("update base salary: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("base salary not found")
	}
	return nil
}

func (r *PostgresEmployeeBaseSalaryRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeleteBaseSalary, id)
	if err != nil {
		return fmt.Errorf("delete base salary: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("base salary not found")
	}
	return nil
}

func baseSalaryModelToEntity(m *EmployeeBaseSalaryModel) *entity.EmployeeBaseSalary {
	return entity.ReconstituteEmployeeBaseSalary(
		m.ID, m.EmployeeID, m.Amount, m.Currency,
		m.EffectiveDate, m.EndDate, m.Notes,
		m.CreatedAt, m.UpdatedAt,
	)
}
