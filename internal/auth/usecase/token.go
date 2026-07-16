package usecase

import "hrms/internal/auth/entity"

type TokenGenerator interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken() (string, error)
}
