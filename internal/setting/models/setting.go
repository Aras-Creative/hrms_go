package models

// UpdateSettingInput is the DTO for updating app settings.
type UpdateSettingInput struct {
	Timezone         string
	CompanyName      string
	CompanyAddress   string
	CompanyLogoID    *string
	WhitelistIPCIDRs []string
}
