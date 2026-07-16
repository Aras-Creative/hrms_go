package repository

import (
	"context"

	"hrms/internal/setting/entity"
)

type SettingRepository interface {
	Find(ctx context.Context) (*entity.Setting, error)
	Upsert(ctx context.Context, s *entity.Setting) error
}
