package adapter

import (
	"context"

	contractUc "hrms/internal/contract/usecase"
	authRepo "hrms/internal/auth/repository"
)

type UserActivatorAdapter struct {
	userRepo authRepo.UserRepository
}

func NewUserActivatorAdapter(userRepo authRepo.UserRepository) *UserActivatorAdapter {
	return &UserActivatorAdapter{userRepo: userRepo}
}

func (a *UserActivatorAdapter) Activate(ctx context.Context, userID string) error {
	return a.userRepo.Activate(ctx, userID)
}

var _ contractUc.UserActivator = (*UserActivatorAdapter)(nil)
