package delivery

import "hrms/internal/auth/models"

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=3,max=100"`
	Role     string `json:"role" validate:"required,oneof=admin"`
	Username string `json:"username" validate:"required,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

func (r *RegisterRequest) ToRegisterInput() *models.RegisterUserInput {
	return &models.RegisterUserInput{
		FullName: r.FullName,
		Role:     r.Role,
		Username: r.Username,
		Password: r.Password,
	}
}

type LoginAdminRequest struct {
	Username string `json:"username" validate:"required,max=255"`
	Password string `json:"password" validate:"required"`
}

func (r *LoginAdminRequest) ToLoginAdminInput() models.LoginAdminInput {
	return models.LoginAdminInput{
		Username: r.Username,
		Password: r.Password,
	}
}

type ChangeNameRequest struct {
	FullName string `json:"full_name" validate:"required,min=3,max=100"`
	Username string `json:"username" validate:"required,max=255"`
}

type RequestChallengeRequest struct {
	Username string `json:"username" validate:"required,max=255"`
}

type DeviceInfo struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

type UserLoginRequest struct {
	Username    string     `json:"username" validate:"required,max=255"`
	ChallengeID string     `json:"challenge_id" validate:"required,uuid"`
	Challenge   string     `json:"challenge" validate:"required"`
	Signature   string     `json:"signature" validate:"required"`
	PublicKey   string     `json:"public_key" validate:"required"`
	Device      DeviceInfo `json:"device" validate:"required"`
}
