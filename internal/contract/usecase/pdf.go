package usecase

import "context"

type PDFRenderer interface {
	Render(ctx context.Context, htmlContent []byte) ([]byte, error)
}
