package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/auth/entity"
)

const (
	queryInsertChallenge = `
		INSERT INTO challenges (id, user_id, challenge_hash, expires_at, is_used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	queryMarkChallengeUsed = `
		UPDATE challenges
		SET is_used = true
		WHERE id = $1 AND is_used = false
	`
)

var (
	queryChallengeByID = `
		SELECT id, user_id, challenge_hash, expires_at, is_used, created_at
		FROM challenges
		WHERE id = $1
	`
)

type PostgresChallengeRepo struct {
	db *sqlx.DB
}

func NewPostgresChallengeRepo(db *sqlx.DB) *PostgresChallengeRepo {
	return &PostgresChallengeRepo{db: db}
}

func (r *PostgresChallengeRepo) CreateChallenge(ctx context.Context, challenge *entity.Challenge) error {
	_, err := r.db.ExecContext(ctx, queryInsertChallenge,
		challenge.ID,
		challenge.UserID,
		challenge.ChallengeHash,
		challenge.ExpiresAt,
		challenge.IsUsed,
		challenge.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}
	return nil
}

func (r *PostgresChallengeRepo) FindByID(ctx context.Context, id string) (*entity.Challenge, error) {
	var c entity.Challenge
	err := r.db.QueryRowxContext(ctx, queryChallengeByID, id).Scan(
		&c.ID, &c.UserID, &c.ChallengeHash, &c.ExpiresAt, &c.IsUsed, &c.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find challenge: %w", err)
	}
	return &c, nil
}

func (r *PostgresChallengeRepo) UpdateChallenge(ctx context.Context, challenge *entity.Challenge) error {
	result, err := r.db.ExecContext(ctx, queryMarkChallengeUsed, challenge.ID)
	if err != nil {
		return fmt.Errorf("failed to update challenge: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("challenge already used")
	}
	return nil
}
