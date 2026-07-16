package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/auth/entity"
)

const (
	queryInsertUser = `
		INSERT INTO users (id, username, password_hash, full_name, is_active, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	querySelectUser = `
		SELECT id, username, password_hash, full_name, is_active, role, created_at, updated_at
		FROM users
	`
	queryUpdateUser = `
		UPDATE users
		SET username = $1, password_hash = $2, full_name = $3,
		    is_active = $4, role = $5, updated_at = $6
		WHERE id = $7
	`
	queryDeleteUser = `DELETE FROM users WHERE id = $1`
	queryExistsUser = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
)

var (
	queryUserByID       = querySelectUser + ` WHERE id = $1`
	queryUserByUsername = querySelectUser + ` WHERE username = $1`
	queryUsersPaginated = querySelectUser + ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
)

func scanUser(scanner sqlx.ColScanner, user *entity.User) error {
	var role string
	err := scanner.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FullName,
		&user.IsActive,
		&role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	user.Role = entity.Role(role)
	return nil
}

type PostgresUserRepo struct {
	db *sqlx.DB
}

func NewPostgresUserRepo(db *sqlx.DB) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) Create(ctx context.Context, user *entity.User) error {
	_, err := r.db.ExecContext(ctx, queryInsertUser,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.FullName,
		user.IsActive,
		string(user.Role),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := scanUser(r.db.QueryRowxContext(ctx, queryUserByID, id), &user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &user, nil
}

func (r *PostgresUserRepo) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := scanUser(r.db.QueryRowxContext(ctx, queryUserByUsername, username), &user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	return &user, nil
}

func (r *PostgresUserRepo) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	rows, err := r.db.QueryxContext(ctx, queryUsersPaginated, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := scanUser(rows, &user); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return users, nil
}

func (r *PostgresUserRepo) Update(ctx context.Context, user *entity.User) error {
	result, err := r.db.ExecContext(ctx, queryUpdateUser,
		user.Username,
		user.PasswordHash,
		user.FullName,
		user.IsActive,
		string(user.Role),
		time.Now(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user with id %s not found", user.ID)
	}
	return nil
}

func (r *PostgresUserRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, queryDeleteUser, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}
	return nil
}

func (r *PostgresUserRepo) Exists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowxContext(ctx, queryExistsUser, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresUserRepo) Deactivate(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) Activate(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_active = true, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) FindAdminIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.db.SelectContext(ctx, &ids, `SELECT id FROM users WHERE role IN ('admin', 'super') AND is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin ids: %w", err)
	}
	return ids, nil
}
