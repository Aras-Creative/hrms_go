package delivery

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	response "hrms/internal/pkg/api"
	"hrms/internal/setting/models"
	"hrms/internal/setting/usecase"
)

type SettingHandler struct {
	uc *usecase.SettingUsecase
}

func NewSettingHandler(uc *usecase.SettingUsecase) *SettingHandler {
	return &SettingHandler{uc: uc}
}

func (h *SettingHandler) Get(c fiber.Ctx) error {
	s, err := h.uc.Get(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}

	var logoURL *string
	if s.CompanyLogoID != nil && *s.CompanyLogoID != "" {
		url, err := h.uc.LogoResolver().ResolveURL(c.RequestCtx(), *s.CompanyLogoID)
		if err != nil {
			return response.Error(c, fmt.Errorf("resolve logo: %w", err))
		}
		logoURL = &url
	}

	return response.OK(c, toResponse(s, logoURL))
}

func (h *SettingHandler) Update(c fiber.Ctx) error {
	var req UpdateSettingRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.UpdateSettingInput{
		Timezone:         req.Timezone,
		CompanyName:      req.CompanyName,
		CompanyAddress:   req.CompanyAddress,
		CompanyLogoID:    req.CompanyLogoID,
		WhitelistIPCIDRs: req.WhitelistIPCIDRs,
	}

	s, err := h.uc.Update(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}

	h.uc.ApplyTimezone(s)

	var logoURL *string
	if s.CompanyLogoID != nil && *s.CompanyLogoID != "" {
		url, err := h.uc.LogoResolver().ResolveURL(c.RequestCtx(), *s.CompanyLogoID)
		if err != nil {
			return response.Error(c, fmt.Errorf("resolve logo: %w", err))
		}
		logoURL = &url
	}

	return response.OK(c, toResponse(s, logoURL))
}
