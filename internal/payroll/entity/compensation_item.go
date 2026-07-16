package entity

import (
	"time"

	"github.com/google/uuid"
)

type CompensationItem struct {
	ID          string
	Name        string
	ItemType    CompensationItemType
	Description string
	IsActive    bool
	IsTaxable   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewCompensationItem(
	name string,
	itemType CompensationItemType,
	description string,
	isTaxable bool,
) *CompensationItem {
	now := time.Now()
	return &CompensationItem{
		ID:          uuid.New().String(),
		Name:        name,
		ItemType:    itemType,
		Description: description,
		IsActive:    true,
		IsTaxable:   isTaxable,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func ReconstituteCompensationItem(
	id string,
	name string,
	itemType string,
	description string,
	isActive bool,
	isTaxable bool,
	createdAt time.Time,
	updatedAt time.Time,
) *CompensationItem {
	return &CompensationItem{
		ID:          id,
		Name:        name,
		ItemType:    CompensationItemType(itemType),
		Description: description,
		IsActive:    isActive,
		IsTaxable:   isTaxable,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
