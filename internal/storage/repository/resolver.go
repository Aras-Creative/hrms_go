package repository

import (
	"strings"

	appConfig "hrms/internal/pkg/config"
)

type cdnResolver struct {
	baseURL string
}

func NewCDNURLResolver(cfg *appConfig.S3Config) URLResolver {
	base := strings.TrimRight(cfg.PublicURLBase, "/")
	return &cdnResolver{baseURL: base}
}

func (r *cdnResolver) PublicURL(storageKey string) string {
	if r.baseURL == "" || storageKey == "" {
		return ""
	}
	return r.baseURL + "/" + strings.TrimLeft(storageKey, "/")
}
