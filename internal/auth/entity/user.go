package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
	FullName     string
	IsActive     bool
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func newUser(username string, passwordHash string, fullName string, role Role) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func NewUser(username string, passwordHash string, fullName string) *User {
	return newUser(username, passwordHash, fullName, RoleUser)
}

func NewUserWithRole(username string, passwordHash string, fullName string, role Role) *User {
	return newUser(username, passwordHash, fullName, role)
}

func NewAdmin(username string, passwordHash string, fullName string) *User {
	return newUser(username, passwordHash, fullName, RoleAdmin)
}

func NewSuperAdmin(username string, passwordHash string, fullName string) *User {
	return newUser(username, passwordHash, fullName, RoleSuper)
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleSuper || u.Role == RoleAdmin
}

func (u *User) IsUser() bool {
	return u.Role == RoleUser
}

func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
}

func ReconstituteUser(
	id string,
	username string,
	passwordHash string,
	fullName string,
	isActive bool,
	role Role,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsActive:     isActive,
		Role:         role,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func (u *User) ChangeName(username, fullName string) {
	u.FullName = fullName
	u.Username = username
	u.UpdatedAt = time.Now()
}
