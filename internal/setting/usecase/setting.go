package usecase

import (
	"context"
	"fmt"
	"sync"

	errors "hrms/internal/pkg/apperror"
	"hrms/internal/pkg/timeutil"
	"hrms/internal/setting/entity"
	"hrms/internal/setting/models"
	"hrms/internal/setting/repository"
)

type SettingUsecase struct {
	repo         repository.SettingRepository
	logoResolver LogoResolver
	mu           sync.RWMutex
	cached       *entity.Setting
}

func NewSettingUsecase(repo repository.SettingRepository, logoResolver LogoResolver) *SettingUsecase {
	return &SettingUsecase{repo: repo, logoResolver: logoResolver}
}

// LogoResolver exposes the resolver for the delivery layer to use.
func (uc *SettingUsecase) LogoResolver() LogoResolver {
	return uc.logoResolver
}

// Get returns the singleton setting from cache, falling back to the database.
func (uc *SettingUsecase) Get(ctx context.Context) (*entity.Setting, error) {
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()

	if cached != nil {
		return cached, nil
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	// Double-check after acquiring write lock
	if uc.cached != nil {
		return uc.cached, nil
	}

	s, err := uc.repo.Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	if s == nil {
		s = entity.NewSetting("Asia/Jakarta", "", "", nil, []string{})
	}

	uc.cached = s
	return s, nil
}

// Update persists the setting and refreshes the in-memory cache.
func (uc *SettingUsecase) Update(ctx context.Context, input models.UpdateSettingInput) (*entity.Setting, error) {
	if len(input.WhitelistIPCIDRs) > 0 {
		for _, cidr := range input.WhitelistIPCIDRs {
			if cidr == "" {
				return nil, errors.NewInvalidInput("CIDR must not be empty")
			}
		}
	}

	s, err := uc.repo.Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("find existing settings for update: %w", err)
	}

	if s == nil {
		s = entity.NewSetting(
			input.Timezone,
			input.CompanyName,
			input.CompanyAddress,
			input.CompanyLogoID,
			input.WhitelistIPCIDRs,
		)
	} else {
		s.Update(
			input.Timezone,
			input.CompanyName,
			input.CompanyAddress,
			input.CompanyLogoID,
			input.WhitelistIPCIDRs,
		)
	}

	if err := uc.repo.Upsert(ctx, s); err != nil {
		return nil, fmt.Errorf("upsert settings: %w", err)
	}

	uc.mu.Lock()
	uc.cached = s
	uc.mu.Unlock()

	return s, nil
}

// ApplyTimezone sets the global default timezone used by timeutil.
func (uc *SettingUsecase) ApplyTimezone(setting *entity.Setting) {
	if setting.Timezone != "" {
		timeutil.SetDefaultTimezone(setting.Timezone)
	}
}
