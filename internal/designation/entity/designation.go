package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Designation struct {
	ID        string
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDesignation(name, code string) *Designation {
	now := time.Now()
	return &Designation{
		ID:        uuid.New().String(),
		Code:      strings.ToUpper(code),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func ReconstituteDesignation(id, code, name string, createdAt, updatedAt time.Time) *Designation {
	return &Designation{
		ID:        id,
		Code:      code,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (d *Designation) Rename(name string) {
	d.Name = name
	d.UpdatedAt = time.Now()
}
