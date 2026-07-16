package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/setting/entity"
)

const (
	findSettingsQuery = `
		SELECT id, timezone, company_name, company_address,
		       company_logo_id, whitelist_ip_cidrs,
		       created_at, updated_at
		FROM settings
		LIMIT 1`

	upsertSettingsQuery = `
		INSERT INTO settings (id, timezone, company_name, company_address, company_logo_id, whitelist_ip_cidrs, updated_at)
		VALUES (:id, :timezone, :company_name, :company_address, :company_logo_id, :whitelist_ip_cidrs, :updated_at)
		ON CONFLICT (id) DO UPDATE SET
			timezone           = EXCLUDED.timezone,
			company_name       = EXCLUDED.company_name,
			company_address    = EXCLUDED.company_address,
			company_logo_id    = EXCLUDED.company_logo_id,
			whitelist_ip_cidrs = EXCLUDED.whitelist_ip_cidrs,
			updated_at         = EXCLUDED.updated_at`
)

type PostgresSettingRepo struct {
	db *sqlx.DB
}

func NewPostgresSettingRepo(db *sqlx.DB) *PostgresSettingRepo {
	return &PostgresSettingRepo{db: db}
}

func (r *PostgresSettingRepo) Find(ctx context.Context) (*entity.Setting, error) {
	var m SettingModel
	if err := r.db.GetContext(ctx, &m, findSettingsQuery); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find settings: %w", err)
	}
	return modelToEntity(&m), nil
}

func (r *PostgresSettingRepo) Upsert(ctx context.Context, s *entity.Setting) error {
	m := entityToModel(s)
	_, err := r.db.NamedExecContext(ctx, upsertSettingsQuery, m)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

func modelToEntity(m *SettingModel) *entity.Setting {
	var createdAt, updatedAt time.Time
	createdAt, _ = time.Parse("2006-01-02T15:04:05Z", m.CreatedAt)
	updatedAt, _ = time.Parse("2006-01-02T15:04:05Z", m.UpdatedAt)
	if createdAt.IsZero() {
		createdAt, _ = time.Parse("2006-01-02T15:04:05.999999Z", m.CreatedAt)
		updatedAt, _ = time.Parse("2006-01-02T15:04:05.999999Z", m.UpdatedAt)
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
		updatedAt = time.Now()
	}

	cidrs := []string(m.WhitelistIPCIDRs)
	if cidrs == nil {
		cidrs = []string{}
	}

	return entity.ReconstituteSetting(
		m.ID,
		m.Timezone,
		m.CompanyName,
		m.CompanyAddress,
		m.CompanyLogoID,
		cidrs,
		createdAt,
		updatedAt,
	)
}

func entityToModel(e *entity.Setting) *SettingModel {
	return &SettingModel{
		ID:               e.ID,
		Timezone:         e.Timezone,
		CompanyName:      e.CompanyName,
		CompanyAddress:   e.CompanyAddress,
		CompanyLogoID:    e.CompanyLogoID,
		WhitelistIPCIDRs: StringSlice(e.WhitelistIPCIDRs),
		UpdatedAt:        e.UpdatedAt.Format(time.RFC3339),
	}
}
