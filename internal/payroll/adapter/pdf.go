package adapter

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	payrollUc "hrms/internal/payroll/usecase"
)

type ChromedpRenderer struct{}

func NewChromedpRenderer() *ChromedpRenderer {
	return &ChromedpRenderer{}
}

func (r *ChromedpRenderer) Render(ctx context.Context, htmlContent []byte) ([]byte, error) {
	allocCtx, cancel := chromedp.NewExecAllocator(ctx,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	defer cancel()

	ctx2, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString(htmlContent)

	var pdfBuf []byte
	err := chromedp.Run(ctx2,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
		pdfBuf, _, err = page.PrintToPDF().
			WithPrintBackground(true).
			WithMarginTop(0).
			WithMarginBottom(0).
			WithMarginLeft(0).
			WithMarginRight(0).
			WithPreferCSSPageSize(true).
			Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp pdf: %w", err)
	}

	return pdfBuf, nil
}

var _ payrollUc.PDFRenderer = (*ChromedpRenderer)(nil)
