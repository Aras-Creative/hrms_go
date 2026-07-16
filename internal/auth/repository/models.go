package repository

import (
	"time"
)

// UserModel is the database-mapped struct for the users table.
type UserModel struct {
	ID           string    `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	FullName     string    `db:"full_name"`
	IsActive     bool      `db:"is_active"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// SessionModel is the database-mapped struct for the sessions table.
type SessionModel struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	DeviceID     *string   `db:"device_id"`
	RefreshToken string    `db:"refresh_token"`
	IsActive     bool      `db:"is_active"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}
