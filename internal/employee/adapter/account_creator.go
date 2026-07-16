package adapter

import (
	"context"

	authModels "hrms/internal/auth/models"
	authUc "hrms/internal/auth/usecase"
	emplUc "hrms/internal/employee/usecase"
)

type AccountCreatorAdapter struct {
	uc *authUc.AuthUsecase
}

func NewAccountCreatorAdapter(uc *authUc.AuthUsecase) *AccountCreatorAdapter {
	return &AccountCreatorAdapter{uc: uc}
}

func (a *AccountCreatorAdapter) CreateUser(ctx context.Context, username, fullName string) (string, error) {
	user, err := a.uc.Register(ctx, &authModels.RegisterUserInput{
		Username: username,
		Password: "",
		FullName: fullName,
		Role:     "user",
	})
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

var _ emplUc.AccountCreator = (*AccountCreatorAdapter)(nil)
