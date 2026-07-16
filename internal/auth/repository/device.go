package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/auth/entity"
)

const (
	queryInsertDevice = `
		INSERT INTO devices (id, user_id, public_key, platform, user_agent, is_active, last_used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	queryUpdateDevice = `
		UPDATE devices
		SET public_key = $1, platform = $2, user_agent = $3, is_active = $4, last_used_at = $5
		WHERE id = $6
	`
)

var (
	queryDeviceByUserID = `
		SELECT id, user_id, public_key, platform, user_agent, is_active, last_used_at, created_at
		FROM devices
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	queryRevokeDeviceByUserID = `UPDATE devices SET is_active = false WHERE user_id = $1`
)

type PostgresDeviceRepo struct {
	db *sqlx.DB
}

func NewPostgresDeviceRepo(db *sqlx.DB) *PostgresDeviceRepo {
	return &PostgresDeviceRepo{db: db}
}

func (r *PostgresDeviceRepo) CreateDevice(ctx context.Context, device *entity.Device) error {
	_, err := r.db.ExecContext(ctx, queryInsertDevice,
		device.ID,
		device.UserID,
		device.PublicKey,
		device.Platform,
		device.UserAgent,
		device.IsActive,
		device.LastUsedAt,
		device.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}
	return nil
}

func (r *PostgresDeviceRepo) FindByUserID(ctx context.Context, userID string) (*entity.Device, error) {
	var d entity.Device
	err := r.db.QueryRowxContext(ctx, queryDeviceByUserID, userID).Scan(
		&d.ID, &d.UserID, &d.PublicKey, &d.Platform, &d.UserAgent, &d.IsActive, &d.LastUsedAt, &d.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find device: %w", err)
	}
	return &d, nil
}

func (r *PostgresDeviceRepo) RevokeDeviceByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, queryRevokeDeviceByUserID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke device: %w", err)
	}
	return nil
}

func (r *PostgresDeviceRepo) UpdateDevice(ctx context.Context, device *entity.Device) error {
	_, err := r.db.ExecContext(ctx, queryUpdateDevice,
		device.PublicKey,
		device.Platform,
		device.UserAgent,
		device.IsActive,
		device.LastUsedAt,
		device.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	return nil
}
