package entity

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type DeductionType struct {
	ID            string
	Name          string
	Slug          string
	Description   string
	DeductionType DeductionCalcType
	DefaultValue  float64
	IsActive      bool
	IsMandatory   bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewDeductionType(
	name string,
	slug string,
	description string,
	deductionType DeductionCalcType,
	defaultValue float64,
	isMandatory bool,
) *DeductionType {
	now := time.Now()
	return &DeductionType{
		ID:            uuid.New().String(),
		Name:          name,
		Slug:          slug,
		Description:   description,
		DeductionType: deductionType,
		DefaultValue:  defaultValue,
		IsActive:      true,
		IsMandatory:   isMandatory,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func ReconstituteDeductionType(
	id string,
	name string,
	slug string,
	description string,
	deductionType string,
	defaultValue float64,
	isActive bool,
	isMandatory bool,
	createdAt time.Time,
	updatedAt time.Time,
) *DeductionType {
	return &DeductionType{
		ID:            id,
		Name:          name,
		Slug:          slug,
		Description:   description,
		DeductionType: DeductionCalcType(deductionType),
		DefaultValue:  defaultValue,
		IsActive:      isActive,
		IsMandatory:   isMandatory,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

// Calculate returns deduction amount in cents.
// For percentage: salaryCents * rate / 100
// For fixed (absent): absentDays * dailyRate * rate / 100, where dailyRate = salaryCents / workingDays
func (dt *DeductionType) Calculate(salaryCents int64, absentDays int, workingDays int) int64 {
	if dt.DeductionType == DeductionCalcPercentage {
		return int64(math.Round(float64(salaryCents) * dt.DefaultValue / 100))
	}
	dailyRate := float64(salaryCents) / float64(workingDays)
	return int64(math.Round(float64(absentDays) * dailyRate * dt.DefaultValue / 100))
}
