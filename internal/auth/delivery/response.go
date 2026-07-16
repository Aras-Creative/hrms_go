package delivery

import (
	"hrms/internal/auth/entity"
	"hrms/internal/auth/models"
	"time"
)

type RegisterResponse struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewRegisterResponse(user *entity.User) *RegisterResponse {
	if user == nil {
		return nil
	}
	return &RegisterResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

type LoginAdminResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewLoginAdminResponse(result *models.LoginResult) *LoginAdminResponse {
	if result == nil || result.User == nil {
		return nil
	}
	return &LoginAdminResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: &UserInfo{
			ID:        result.User.ID,
			FullName:  result.User.FullName,
			Username:  result.User.Username,
			Role:      string(result.User.Role),
			IsActive:  result.User.IsActive,
			CreatedAt: result.User.CreatedAt,
			UpdatedAt: result.User.UpdatedAt,
		},
	}
}

type RefreshTokenResponse = LoginAdminResponse

type ChallengeResponse struct {
	ChallengeID string    `json:"challenge_id"`
	Challenge   string    `json:"challenge"`
	ExpiresAt   time.Time `json:"expires_at"`
}
