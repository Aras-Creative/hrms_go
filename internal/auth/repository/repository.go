package repository

import (
	"context"

	"hrms/internal/auth/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Deactivate(ctx context.Context, userID string) error
	Activate(ctx context.Context, userID string) error
	Exists(ctx context.Context, username string) (bool, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindAdminIDs(ctx context.Context) ([]string, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *entity.Session) error
	UpdateSession(ctx context.Context, session *entity.Session) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (*entity.Session, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error)
	RevokeAllByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type ChallengeRepository interface {
	CreateChallenge(ctx context.Context, challenge *entity.Challenge) error
	FindByID(ctx context.Context, id string) (*entity.Challenge, error)
	UpdateChallenge(ctx context.Context, challenge *entity.Challenge) error
}

type DeviceRepository interface {
	CreateDevice(ctx context.Context, device *entity.Device) error
	FindByUserID(ctx context.Context, userID string) (*entity.Device, error)
	UpdateDevice(ctx context.Context, device *entity.Device) error
	RevokeDeviceByUserID(ctx context.Context, userID string) error
}
