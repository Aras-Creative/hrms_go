package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WorkingType string

const (
	WorkingTypeFixed   WorkingType = "fixed"
	WorkingTypeDynamic WorkingType = "dynamic"
	WorkingTypeOff     WorkingType = "off"
)

var validWorkingTypes = []WorkingType{WorkingTypeFixed, WorkingTypeDynamic, WorkingTypeOff}

func ParseWorkingType(s string) (WorkingType, error) {
	wt := WorkingType(s)
	if !wt.IsValid() {
		return "", fmt.Errorf("invalid working type: %s", s)
	}
	return wt, nil
}

func (wt WorkingType) IsValid() bool {
	for _, valid := range validWorkingTypes {
		if wt == valid {
			return true
		}
	}
	return false
}

type DayOfWeek int

const (
	Sunday    DayOfWeek = 0
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
	Saturday  DayOfWeek = 6
)

var validDaysOfWeek = []DayOfWeek{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}

func ParseDayOfWeek(v int) (DayOfWeek, error) {
	d := DayOfWeek(v)
	if !d.IsValid() {
		return 0, fmt.Errorf("invalid day of week: %d", v)
	}
	return d, nil
}

func (d DayOfWeek) IsValid() bool {
	for _, valid := range validDaysOfWeek {
		if d == valid {
			return true
		}
	}
	return false
}

type WorkingPatternDetail struct {
	ID               string
	WorkingPatternID string
	DayOfWeek        DayOfWeek

	Type WorkingType

	StartTime *string
	EndTime   *string
}

type WorkingPattern struct {
	ID          string
	Name        string
	Description *string
	IsActive    bool
	Details     []WorkingPatternDetail
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewWorkingPattern(name string, description *string, details []WorkingPatternDetail) *WorkingPattern {
	now := time.Now()
	id := uuid.New().String()
	for i := range details {
		details[i].ID = uuid.New().String()
		details[i].WorkingPatternID = id
	}
	return &WorkingPattern{
		ID:          id,
		Name:        name,
		Description: description,
		IsActive:    true,
		Details:     details,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func ReconstituteWorkingPattern(
	id, name string,
	description *string,
	isActive bool,
	details []WorkingPatternDetail,
	createdAt, updatedAt time.Time,
) *WorkingPattern {
	return &WorkingPattern{
		ID:          id,
		Name:        name,
		Description: description,
		IsActive:    isActive,
		Details:     details,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func (w *WorkingPattern) Rename(name string) {
	w.Name = name
	w.UpdatedAt = time.Now()
}

func (w *WorkingPattern) Disable() error {
	if !w.IsActive {
		return fmt.Errorf("pattern is already disabled")
	}
	w.IsActive = false
	w.UpdatedAt = time.Now()
	return nil
}
