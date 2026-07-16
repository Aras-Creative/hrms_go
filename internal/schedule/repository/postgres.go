package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/schedule/entity"
)

const (
	queryInsertWorkPattern = `
		INSERT INTO work_patterns (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	queryInsertWorkPatternDetail = `
		INSERT INTO work_pattern_details (id, work_pattern_id, day_of_week, working_type, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	querySelectWorkPattern = `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM work_patterns
	`

	querySelectDetailsByPatternID = `
		SELECT id, work_pattern_id, day_of_week, working_type, start_time, end_time
		FROM work_pattern_details
		WHERE work_pattern_id = $1
		ORDER BY day_of_week ASC
	`

	querySelectDetailsByPatternIDs = `
		SELECT id, work_pattern_id, day_of_week, working_type, start_time, end_time
		FROM work_pattern_details
		WHERE work_pattern_id = ANY($1)
		ORDER BY work_pattern_id, day_of_week ASC
	`

	queryDeleteDetailsByPatternID = `
		DELETE FROM work_pattern_details WHERE work_pattern_id = $1
	`

	queryUpdateWorkPattern = `
		UPDATE work_patterns SET name = $1, description = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`
)

var (
	queryWorkPatternByID       = querySelectWorkPattern + ` WHERE id = $1`
	queryWorkPatternsAll       = querySelectWorkPattern + ` ORDER BY name ASC`
	queryWorkPatternsAllActive = querySelectWorkPattern + ` WHERE is_active = true ORDER BY name ASC`
)

type PostgresWorkPatternRepo struct {
	db *sqlx.DB
}

func NewPostgresWorkPatternRepo(db *sqlx.DB) *PostgresWorkPatternRepo {
	return &PostgresWorkPatternRepo{db: db}
}

func (r *PostgresWorkPatternRepo) Create(ctx context.Context, wp *entity.WorkingPattern) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, queryInsertWorkPattern,
		wp.ID, wp.Name, wp.Description, wp.IsActive, wp.CreatedAt, wp.UpdatedAt,
	); err != nil {
		return fmt.Errorf("failed to create work pattern: %w", err)
	}

	for _, d := range wp.Details {
		if _, err := tx.ExecContext(ctx, queryInsertWorkPatternDetail,
			d.ID, d.WorkingPatternID, int(d.DayOfWeek), string(d.Type), d.StartTime, d.EndTime,
		); err != nil {
			return fmt.Errorf("failed to create work pattern detail: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresWorkPatternRepo) FindByID(ctx context.Context, id string) (*entity.WorkingPattern, error) {
	var m WorkPatternModel
	err := r.db.QueryRowxContext(ctx, queryWorkPatternByID, id).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find work pattern: %w", err)
	}

	details, err := r.findDetailsByPatternID(ctx, id)
	if err != nil {
		return nil, err
	}

	return modelToWorkPattern(&m, details), nil
}

func (r *PostgresWorkPatternRepo) FindAll(ctx context.Context) ([]*entity.WorkingPattern, error) {
	return r.findAll(ctx, queryWorkPatternsAll)
}

func (r *PostgresWorkPatternRepo) FindAllActive(ctx context.Context) ([]*entity.WorkingPattern, error) {
	return r.findAll(ctx, queryWorkPatternsAllActive)
}

func (r *PostgresWorkPatternRepo) Update(ctx context.Context, wp *entity.WorkingPattern) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, queryUpdateWorkPattern,
		wp.Name, wp.Description, wp.IsActive, wp.UpdatedAt, wp.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update work pattern: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("work pattern not found")
	}

	if _, err := tx.ExecContext(ctx, queryDeleteDetailsByPatternID, wp.ID); err != nil {
		return fmt.Errorf("failed to delete old details: %w", err)
	}

	for _, d := range wp.Details {
		if _, err := tx.ExecContext(ctx, queryInsertWorkPatternDetail,
			d.ID, d.WorkingPatternID, int(d.DayOfWeek), string(d.Type), d.StartTime, d.EndTime,
		); err != nil {
			return fmt.Errorf("failed to insert detail: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresWorkPatternRepo) findAll(ctx context.Context, query string) ([]*entity.WorkingPattern, error) {
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list work patterns: %w", err)
	}
	defer rows.Close()

	var models []WorkPatternModel
	for rows.Next() {
		var m WorkPatternModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("failed to scan work pattern: %w", err)
		}
		models = append(models, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, nil
	}

	ids := make([]string, len(models))
	for i, m := range models {
		ids[i] = m.ID
	}

	detailRows, err := r.db.QueryxContext(ctx, querySelectDetailsByPatternIDs, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to list details: %w", err)
	}
	defer detailRows.Close()

	detailsByPattern := make(map[string][]WorkPatternDetailModel)
	for detailRows.Next() {
		var dm WorkPatternDetailModel
		if err := detailRows.StructScan(&dm); err != nil {
			return nil, fmt.Errorf("failed to scan detail: %w", err)
		}
		detailsByPattern[dm.WorkingPatternID] = append(detailsByPattern[dm.WorkingPatternID], dm)
	}

	list := make([]*entity.WorkingPattern, 0, len(models))
	for i := range models {
		list = append(list, modelToWorkPattern(&models[i], detailsByPattern[models[i].ID]))
	}
	return list, nil
}

func (r *PostgresWorkPatternRepo) findDetailsByPatternID(ctx context.Context, patternID string) ([]WorkPatternDetailModel, error) {
	rows, err := r.db.QueryxContext(ctx, querySelectDetailsByPatternID, patternID)
	if err != nil {
		return nil, fmt.Errorf("failed to list details: %w", err)
	}
	defer rows.Close()

	var details []WorkPatternDetailModel
	for rows.Next() {
		var dm WorkPatternDetailModel
		if err := rows.StructScan(&dm); err != nil {
			return nil, fmt.Errorf("failed to scan detail: %w", err)
		}
		details = append(details, dm)
	}
	return details, rows.Err()
}

func modelToWorkPattern(m *WorkPatternModel, details []WorkPatternDetailModel) *entity.WorkingPattern {
	d := make([]entity.WorkingPatternDetail, 0, len(details))
	for _, dm := range details {
		d = append(d, entity.WorkingPatternDetail{
			ID:               dm.ID,
			WorkingPatternID: dm.WorkingPatternID,
			DayOfWeek:        entity.DayOfWeek(dm.DayOfWeek),
			Type:             entity.WorkingType(dm.Type),
			StartTime:        dm.StartTime,
			EndTime:          dm.EndTime,
		})
	}
	return entity.ReconstituteWorkingPattern(
		m.ID, m.Name, m.Description, m.IsActive, d, m.CreatedAt, m.UpdatedAt,
	)
}
