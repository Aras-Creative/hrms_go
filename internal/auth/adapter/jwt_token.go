package adapter

import (
	"hrms/internal/auth/entity"
	pkgJWT "hrms/internal/pkg/jwt"
)

type JWTAdapter struct {
	jwt *pkgJWT.Service
}

func NewJWTAdapter(jwt *pkgJWT.Service) *JWTAdapter {
	return &JWTAdapter{
		jwt: jwt,
	}
}

func (a *JWTAdapter) GenerateAccessToken(user *entity.User) (string, error) {
	return a.jwt.GenerateAccessToken(user.ID, user.Role.String())
}

func (a *JWTAdapter) GenerateRefreshToken() (string, error) {
	return a.jwt.GenerateRefreshToken()
}
