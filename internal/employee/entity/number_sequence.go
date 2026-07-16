package entity

import "time"

type NumberSequence struct {
	DesignationCode string
	Prefix          string
	LastSequence    int
	UpdatedAt       time.Time
}

func NewNumberSequence(designationCode, prefix string) *NumberSequence {
	return &NumberSequence{
		DesignationCode: designationCode,
		Prefix:          prefix,
		LastSequence:    0,
		UpdatedAt:       time.Now(),
	}
}

func ReconstituteNumberSequence(designationCode, prefix string, lastSequence int, updatedAt time.Time) *NumberSequence {
	return &NumberSequence{
		DesignationCode: designationCode,
		Prefix:          prefix,
		LastSequence:    lastSequence,
		UpdatedAt:       updatedAt,
	}
}
