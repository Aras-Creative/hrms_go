package adapter

import (
	"context"

	settingUc "hrms/internal/setting/usecase"
	payrollUc "hrms/internal/payroll/usecase"
)

type CompanySettingsProviderAdapter struct {
	settingUC *settingUc.SettingUsecase
}

func NewCompanySettingsProviderAdapter(settingUC *settingUc.SettingUsecase) *CompanySettingsProviderAdapter {
	return &CompanySettingsProviderAdapter{settingUC: settingUC}
}

func (a *CompanySettingsProviderAdapter) GetCompanySettings(ctx context.Context) (companyName, companyAddress, logoURL string, err error) {
	s, err := a.settingUC.Get(ctx)
	if err != nil {
		return "", "", "", err
	}
	if s == nil {
		return "", "", "", nil
	}

	companyName = s.CompanyName
	companyAddress = s.CompanyAddress

	if s.CompanyLogoID != nil && *s.CompanyLogoID != "" {
		url, err := a.settingUC.LogoResolver().ResolveURL(ctx, *s.CompanyLogoID)
		if err != nil {
			return "", "", "", err
		}
		logoURL = url
	}

	return companyName, companyAddress, logoURL, nil
}

var _ payrollUc.CompanySettingsProvider = (*CompanySettingsProviderAdapter)(nil)
