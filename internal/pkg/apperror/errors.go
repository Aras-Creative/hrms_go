package errors

import "net/http"

type DomainError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *DomainError) Error() string {
	return e.Message
}

var (
	ErrNotFound           = &DomainError{Code: "NOT_FOUND", Message: "resource not found", HTTPStatus: http.StatusNotFound}
	ErrAlreadyExists      = &DomainError{Code: "ALREADY_EXISTS", Message: "resource already exists", HTTPStatus: http.StatusConflict}
	ErrInvalidInput       = &DomainError{Code: "INVALID_INPUT", Message: "invalid input", HTTPStatus: http.StatusBadRequest}
	ErrInternal           = &DomainError{Code: "INTERNAL", Message: "internal server error", HTTPStatus: http.StatusInternalServerError}
	ErrUnauthorized       = &DomainError{Code: "UNAUTHORIZED", Message: "unauthorized", HTTPStatus: http.StatusUnauthorized}
	ErrInvalidCredentials = &DomainError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials", HTTPStatus: http.StatusUnauthorized}
	ErrSessionExpired     = &DomainError{Code: "SESSION_EXPIRED", Message: "session expired", HTTPStatus: http.StatusUnauthorized}
	ErrDeviceRevoked      = &DomainError{Code: "DEVICE_REVOKED", Message: "device has been revoked by admin", HTTPStatus: http.StatusForbidden}
)

func NewNotFound(msg string) *DomainError {
	return &DomainError{Code: "NOT_FOUND", Message: msg, HTTPStatus: http.StatusNotFound}
}

func NewAlreadyExists(msg string) *DomainError {
	return &DomainError{Code: "ALREADY_EXISTS", Message: msg, HTTPStatus: http.StatusConflict}
}

func NewInvalidInput(msg string) *DomainError {
	return &DomainError{Code: "INVALID_INPUT", Message: msg, HTTPStatus: http.StatusBadRequest}
}

func NewInternal(msg string) *DomainError {
	return &DomainError{Code: "INTERNAL", Message: msg, HTTPStatus: http.StatusInternalServerError}
}

func NewUnauthorized(msg string) *DomainError {
	return &DomainError{Code: "UNAUTHORIZED", Message: msg, HTTPStatus: http.StatusUnauthorized}
}

func NewForbidden(msg string) *DomainError {
	return &DomainError{Code: "FORBIDDEN", Message: msg, HTTPStatus: http.StatusForbidden}
}
