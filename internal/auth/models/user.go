package models

import "hrms/internal/auth/entity"

type RegisterUserInput struct {
	FullName string
	Role     string
	Username string
	Password string
}

type LoginAdminInput struct {
	Username string
	Password string
	DeviceID string
}

type RefreshTokenInput struct {
	RefreshToken string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         *entity.User
	Session      *entity.Session
}

type UserLoginInput struct {
	Username    string
	ChallengeID string
	Challenge   string
	Signature   string
	PublicKey   string
	DeviceName  string
	Platform    string
}
