package delivery

type UpdateSettingRequest struct {
	Timezone         string   `json:"timezone" validate:"required"`
	CompanyName      string   `json:"company_name" validate:"required"`
	CompanyAddress   string   `json:"company_address" validate:"required"`
	CompanyLogoID    *string  `json:"company_logo_id"`
	WhitelistIPCIDRs []string `json:"whitelist_ip_cidrs"`
}
