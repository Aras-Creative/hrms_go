package entity

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID         string
	UserID     string
	PublicKey  string
	Platform   string
	UserAgent  string
	IsActive   bool
	LastUsedAt time.Time
	CreatedAt  time.Time
}

func NewDevice(userID, publicKey, platform, userAgent string) *Device {
	return &Device{
		ID:         uuid.New().String(),
		UserID:     userID,
		PublicKey:  publicKey,
		Platform:   platform,
		UserAgent:  userAgent,
		IsActive:   true,
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}
}

func ReconstituteDevice(
	id, userID, publicKey, platform, userAgent string,
	isActive bool,
	lastUsedAt, createdAt time.Time,
) *Device {
	return &Device{
		ID:         id,
		UserID:     userID,
		PublicKey:  publicKey,
		Platform:   platform,
		UserAgent:  userAgent,
		IsActive:   isActive,
		LastUsedAt: lastUsedAt,
		CreatedAt:  createdAt,
	}
}

func (d *Device) IsBound() bool {
	return d.PublicKey != ""
}

func (d *Device) IsActiveDevice() bool {
	return d.IsActive
}

func (d *Device) Touch() {
	d.LastUsedAt = time.Now()
}

func (d *Device) Revoke() {
	d.IsActive = false
}
