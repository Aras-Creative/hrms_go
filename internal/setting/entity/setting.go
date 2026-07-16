package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Setting holds application-wide configuration managed at runtime.
type Setting struct {
	ID               string
	Timezone         string
	CompanyName      string
	CompanyAddress   string
	CompanyLogoID    *string // FK to documents.id
	WhitelistIPCIDRs []string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewSetting(timezone, companyName, companyAddress string, companyLogoID *string, whitelistIPCIDRs []string) *Setting {
	now := time.Now()
	return &Setting{
		ID:               uuid.New().String(),
		Timezone:         timezone,
		CompanyName:      companyName,
		CompanyAddress:   companyAddress,
		CompanyLogoID:    companyLogoID,
		WhitelistIPCIDRs: whitelistIPCIDRs,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func ReconstituteSetting(
	id, timezone, companyName, companyAddress string,
	companyLogoID *string,
	whitelistIPCIDRs []string,
	createdAt, updatedAt time.Time,
) *Setting {
	return &Setting{
		ID:               id,
		Timezone:         timezone,
		CompanyName:      companyName,
		CompanyAddress:   companyAddress,
		CompanyLogoID:    companyLogoID,
		WhitelistIPCIDRs: whitelistIPCIDRs,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

// Update applies non-zero changes and bumps the timestamp.
func (s *Setting) Update(timezone, companyName, companyAddress string, companyLogoID *string, whitelistIPCIDRs []string) {
	if timezone != "" {
		s.Timezone = timezone
	}
	if companyName != "" {
		s.CompanyName = companyName
	}
	if companyAddress != "" {
		s.CompanyAddress = companyAddress
	}
	s.CompanyLogoID = companyLogoID
	s.WhitelistIPCIDRs = whitelistIPCIDRs
	s.UpdatedAt = time.Now()
}

// ValidateCIDRs checks that each entry is a valid CIDR notation.
func (s *Setting) ValidateCIDRs() error {
	for _, cidr := range s.WhitelistIPCIDRs {
		if cidr == "" {
			return fmt.Errorf("whitelist CIDR must not be empty")
		}
	}
	return nil
}
