package delivery

import (
	"hrms/internal/setting/entity"
)

type SettingResponse struct {
	ID               string   `json:"id"`
	Timezone         string   `json:"timezone"`
	CompanyName      string   `json:"company_name"`
	CompanyAddress   string   `json:"company_address"`
	CompanyLogoURL   *string  `json:"company_logo_url"`
	CompanyLogoID    *string  `json:"company_logo_id"`
	WhitelistIPCIDRs []string `json:"whitelist_ip_cidrs"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

func toResponse(s *entity.Setting, logoURL *string) *SettingResponse {
	return &SettingResponse{
		ID:               s.ID,
		Timezone:         s.Timezone,
		CompanyName:      s.CompanyName,
		CompanyAddress:   s.CompanyAddress,
		CompanyLogoURL:   logoURL,
		CompanyLogoID:    s.CompanyLogoID,
		WhitelistIPCIDRs: s.WhitelistIPCIDRs,
		CreatedAt:        s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
