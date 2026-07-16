package usecase

import "context"

// LogoResolver resolves a document ID to a public CDN URL.
type LogoResolver interface {
	ResolveURL(ctx context.Context, documentID string) (string, error)
}
