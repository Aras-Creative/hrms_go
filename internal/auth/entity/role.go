package entity

import errors "hrms/internal/pkg/apperror"

type Role string

const (
	RoleSuper Role = "super"
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

var ErrInvalidRole = errors.NewInvalidInput("invalid role")

func ParseRole(s string) (Role, error) {
	switch Role(s) {
	case RoleSuper:
		return RoleSuper, nil
	case RoleAdmin:
		return RoleAdmin, nil
	case RoleUser:
		return RoleUser, nil
	default:
		return "", ErrInvalidRole
	}
}

func (r Role) String() string {
	return string(r)
}
