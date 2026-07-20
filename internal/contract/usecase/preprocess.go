package usecase

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"hrms/internal/contract/entity"
)

// resolveBlockContent performs variable substitution on block content strings
// using the render context. Each block's Content field is parsed as a Go template
// and executed against the RenderContext. The result is stored in Resolved.
func resolveBlockContent(blocks []RenderBlock, ctx *RenderContext) {
	for i, b := range blocks {
		if b.Content == "" {
			blocks[i].Resolved = template.HTML(b.Content)
			continue
		}
		sub, err := template.New(fmt.Sprintf("block-%d", i)).Funcs(tmplFuncs).Parse(b.Content)
		if err != nil {
			blocks[i].Resolved = template.HTML(b.Content)
			continue
		}
		var buf bytes.Buffer
		if err := sub.Execute(&buf, ctx); err != nil {
			blocks[i].Resolved = template.HTML(b.Content)
			continue
		}
		blocks[i].Resolved = template.HTML(buf.String())
	}
}

// buildContractRenderData maps a contract entity to the render-safe struct.
func buildContractRenderData(c *entity.Contract, shiftStart, shiftEnd string) entity.ContractRenderData {
	salary := c.Salary
	if salary == "" {
		salary = "0"
	}
	return entity.ContractRenderData{
		Number:           c.Number,
		StartDate:        formatTime(c.StartDate),
		EndDate:          formatTime(c.EndDate),
		Salary:           formatSalaryCurrency(salary),
		DesignationTitle: c.DesignationTitle,
		ShiftStart:       shiftStart,
		ShiftEnd:         shiftEnd,
	}
}

// formatSalaryCurrency formats a numeric string as Indonesian Rupiah currency.
// e.g. "5000000" -> "Rp 5.000.000", "5000000.50" -> "Rp 5.000.000,50"
func formatSalaryCurrency(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "Rp 0"
	}

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}

	// split whole and fractional
	parts := strings.SplitN(s, ".", 2)
	wholeStr := strings.ReplaceAll(parts[0], ",", "")
	fracStr := ""
	if len(parts) > 1 {
		fracStr = strings.TrimRight(parts[1], "0")
	}

	n, err := strconv.ParseInt(wholeStr, 10, 64)
	if err != nil {
		return s
	}

	digits := strconv.FormatInt(n, 10)
	var groups []string
	for len(digits) > 3 {
		groups = append(groups, digits[len(digits)-3:])
		digits = digits[:len(digits)-3]
	}
	groups = append(groups, digits)
	for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
		groups[i], groups[j] = groups[j], groups[i]
	}
	result := "Rp " + strings.Join(groups, ".")

	if fracStr != "" {
		result += "," + fracStr
	}
	if neg {
		result = "-" + result
	}
	return result
}

// buildSigningRenderData maps signing entities to render-safe structs.
func buildSigningRenderData(signings []*entity.ContractSigning) []entity.ContractSigningRenderData {
	var out []entity.ContractSigningRenderData
	for _, s := range signings {
		out = append(out, entity.ContractSigningRenderData{
			Party:           s.Party,
			SignedByName:    s.SignedByName,
			SignedByTitle:   s.SignedByTitle,
			SignatureBase64: s.SignatureBase64,
			Place:           s.Place,
			SignedAt:        formatTime(&s.SignedAt),
		})
	}
	return out
}
