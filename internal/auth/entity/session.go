package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           string
	UserID       string
	DeviceID     string
	RefreshToken string
	IsActive     bool
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

func NewSession(userID, deviceID, refreshToken string, expiresAt time.Time) *Session {
	return &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		DeviceID:     deviceID,
		RefreshToken: refreshToken,
		IsActive:     true,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
}

func ReconstituteSession(
	id, userID, deviceID, refreshToken string,
	isActive bool,
	expiresAt, createdAt time.Time,
) *Session {
	return &Session{
		ID:           id,
		UserID:       userID,
		DeviceID:     deviceID,
		RefreshToken: refreshToken,
		IsActive:     isActive,
		ExpiresAt:    expiresAt,
		CreatedAt:    createdAt,
	}
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) Revoke() {
	s.IsActive = false
}

// BelongsToDevice returns true if the session is bound to the given device ID.
// A session with an empty DeviceID is considered unbound (admin login).
func (s *Session) BelongsToDevice(deviceID string) bool {
	return s.DeviceID != "" && s.DeviceID == deviceID
}

// CheckOneDevicePolicy revokes all sessions that belong to the given device,
// and returns an error if any active session belongs to a different device.
// Sessions with no device binding are left untouched.
func CheckOneDevicePolicy(sessions []*Session, deviceID string) error {
	for _, s := range sessions {
		if s.DeviceID != "" && s.DeviceID != deviceID {
			return fmt.Errorf("active session exists on a different device")
		}
		if s.BelongsToDevice(deviceID) {
			s.Revoke()
		}
	}
	return nil
}
