package adapter

import (
	"context"

	authRepo "hrms/internal/auth/repository"
	leaveUc "hrms/internal/leave/usecase"
)

type UserNameAdapter struct {
	repo authRepo.UserRepository
}

func NewUserNameAdapter(repo authRepo.UserRepository) *UserNameAdapter {
	return &UserNameAdapter{repo: repo}
}

func (a *UserNameAdapter) FindNameByID(ctx context.Context, userID string) (string, error) {
	user, err := a.repo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}
	return user.FullName, nil
}

func (a *UserNameAdapter) FindAdminIDs(ctx context.Context) ([]string, error) {
	return a.repo.FindAdminIDs(ctx)
}

var _ leaveUc.UserNameResolver = (*UserNameAdapter)(nil)
