package models

import "time"

type DeviceInfo struct {
	ID         string
	Platform   string
	Name       string
	IsActive   bool
	LastUsedAt time.Time
	CreatedAt  time.Time
}
