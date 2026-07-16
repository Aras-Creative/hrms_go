package usecase

import (
	"context"
	"fmt"
	"hrms/internal/auth/repository"
	errors "hrms/internal/pkg/apperror"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

type ChangeNameInput struct {
	UserID   string
	FullName string
	Username string
}

func (uc *UserUsecase) ChangeName(ctx context.Context, input *ChangeNameInput) error {
	existing, err := uc.repo.FindByID(ctx, input.UserID)
	if err != nil || existing == nil {
		return errors.NewNotFound("user not found")
	}

	existing.ChangeName(input.Username, input.FullName)

	if err := uc.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update user name: %w", err)
	}

	return nil
}
