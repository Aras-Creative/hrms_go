package usecase

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"
	"time"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/repository"
	errors "hrms/internal/pkg/apperror"
)

var tmplFuncs = template.FuncMap{
	"safeURL": func(s string) template.URL { return template.URL(s) },
	"join":    strings.Join,
	"orderedList": func(items []string) template.HTML {
		return buildOrderedListHTML(items)
	},
	"show": func(value, placeholder string) string {
		if value == "" {
			return placeholder
		}
		return value
	},
}

// buildOrderedListHTML converts a string slice into a complete <ol> with <li> items.
func buildOrderedListHTML(items []string) template.HTML {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<ol>\n")
	for _, item := range items {
		b.WriteString("<li>")
		b.WriteString(template.HTMLEscapeString(item))
		b.WriteString("</li>\n")
	}
	b.WriteString("</ol>")
	return template.HTML(b.String())
}

//go:embed templates/contract.html
var contractTemplate string

//go:embed templates/company_logo.jpeg
var companyLogo []byte

// ShiftTimeFetcher fetches an employee's shift start and end time from their work pattern.
type ShiftTimeFetcher interface {
	FindShiftTimesByEmployeeID(ctx context.Context, employeeID string) (start, end string, err error)
}

type FieldRow struct {
	Label string
	Value string
}

type RenderBlock struct {
	Type     string
	Label    string
	Title    string
	Content  string
	Position string
	Fields   []FieldRow
	Resolved template.HTML
}

type RenderContext struct {
	Employee         entity.EmployeeRenderData
	Contract         entity.ContractRenderData
	Signatory        entity.SignatoryRenderData
	Data             entity.ContractTemplateData
	Signings         []entity.ContractSigningRenderData
	Blocks           []RenderBlock
	CompanyLogo      string
	JobDutiesJoined  string
	InventoryJoined  string
	JobDutiesList    template.HTML
	InventoryList    template.HTML
}

func (r *RenderContext) FirstSig() *entity.ContractSigningRenderData {
	for _, s := range r.Signings {
		if s.Party == "first" {
			return &s
		}
	}
	return nil
}

func (r *RenderContext) SecondSig() *entity.ContractSigningRenderData {
	for _, s := range r.Signings {
		if s.Party == "second" {
			return &s
		}
	}
	return nil
}

type RenderUsecase struct {
	contractRepo repository.ContractRepository
	signingRepo  repository.SigningRepository
	empFetcher   EmployeeFetcher
	shiftFetcher ShiftTimeFetcher
	pdf          PDFRenderer
}

func NewRenderUsecase(contractRepo repository.ContractRepository, signingRepo repository.SigningRepository, empFetcher EmployeeFetcher, shiftFetcher ShiftTimeFetcher, pdf PDFRenderer) *RenderUsecase {
	return &RenderUsecase{contractRepo: contractRepo, signingRepo: signingRepo, empFetcher: empFetcher, shiftFetcher: shiftFetcher, pdf: pdf}
}

func (uc *RenderUsecase) Preview(ctx context.Context, contractID string, signatoryName, signatoryTitle string) ([]byte, error) {
	c, err := uc.contractRepo.FindContractByID(ctx, contractID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find contract", err)
	}
	if c == nil {
		return nil, errors.NewNotFound("contract not found")
	}

	emp, err := uc.empFetcher.FindEmployeeRenderData(ctx, c.EmployeeID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find employee", err)
	}
	if emp == nil {
		return nil, errors.NewNotFound("employee not found")
	}

	signings, _ := uc.signingRepo.FindSigningsByContractID(ctx, contractID)

	return uc.renderAndGeneratePDF(ctx, c, emp, signatoryName, signatoryTitle, signings)
}

func (uc *RenderUsecase) PreviewWithSignings(ctx context.Context, contractID string, signatoryName, signatoryTitle string, signings []*entity.ContractSigning) ([]byte, error) {
	c, err := uc.contractRepo.FindContractByID(ctx, contractID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find contract", err)
	}
	if c == nil {
		return nil, errors.NewNotFound("contract not found")
	}

	emp, err := uc.empFetcher.FindEmployeeRenderData(ctx, c.EmployeeID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find employee", err)
	}
	if emp == nil {
		return nil, errors.NewNotFound("employee not found")
	}

	return uc.renderAndGeneratePDF(ctx, c, emp, signatoryName, signatoryTitle, signings)
}

func (uc *RenderUsecase) renderAndGeneratePDF(ctx context.Context, c *entity.Contract, emp *entity.EmployeeRenderData, signatoryName, signatoryTitle string, signings []*entity.ContractSigning) ([]byte, error) {
	shiftStart, shiftEnd, _ := uc.shiftFetcher.FindShiftTimesByEmployeeID(ctx, c.EmployeeID)

	html, err := uc.renderHTML(c, emp, signatoryName, signatoryTitle, signings, shiftStart, shiftEnd)
	if err != nil {
		return nil, err
	}

	return uc.pdf.Render(ctx, html)
}

func (uc *RenderUsecase) renderHTML(c *entity.Contract, emp *entity.EmployeeRenderData, signatoryName, signatoryTitle string, signings []*entity.ContractSigning, shiftStart, shiftEnd string) ([]byte, error) {
	renderCtx := buildRenderContext(c, emp, signatoryName, signatoryTitle, signings, shiftStart, shiftEnd)

	// Resolve variable placeholders in block content
	resolveBlockContent(renderCtx.Blocks, renderCtx)

	tmpl, err := template.New("contract").Funcs(tmplFuncs).Parse(contractTemplate)
	if err != nil {
		return nil, errors.WrapInternal("failed to parse template", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderCtx); err != nil {
		return nil, errors.WrapInternal("failed to execute template", err)
	}

	return buf.Bytes(), nil
}

func buildRenderContext(c *entity.Contract, emp *entity.EmployeeRenderData, signatoryName, signatoryTitle string, signings []*entity.ContractSigning, shiftStart, shiftEnd string) *RenderContext {
	renderCtx := &RenderContext{
		Employee:        *emp,
		Contract:        buildContractRenderData(c, shiftStart, shiftEnd),
		Signatory:       entity.SignatoryRenderData{Name: signatoryName, Designation: signatoryTitle},
		Data:            c.Data,
		Signings:        buildSigningRenderData(signings),
		CompanyLogo:     base64.StdEncoding.EncodeToString(companyLogo),
		JobDutiesJoined: strings.Join(c.Data.JobDuties, "\n"),
		InventoryJoined: strings.Join(c.Data.InventoryItems, "\n"),
		JobDutiesList:   buildOrderedListHTML(c.Data.JobDuties),
		InventoryList:   buildOrderedListHTML(c.Data.InventoryItems),
	}

	articleNum := 0
	paraNum := 0
	for _, b := range c.Templates.Blocks {
		rb := RenderBlock{
			Type:     b.Type,
			Title:    coalesceStr(b.Title),
			Content:  coalesceStr(b.Content),
			Position: coalesceStr(b.Position),
		}

		switch b.Type {
		case "article":
			articleNum++
			paraNum = 0
			rb.Type = "pasal"
			rb.Label = fmt.Sprintf("PASAL %d", articleNum)
		case "paragraph":
			paraNum++
			rb.Type = "ayat"
			rb.Label = fmt.Sprintf("AYAT %d", paraNum)
		case "second_party":
			rb.Fields = buildFields(c.Data.EmployeeFields, emp)
		}

		renderCtx.Blocks = append(renderCtx.Blocks, rb)
	}

	if len(renderCtx.Blocks) == 0 {
		renderCtx.Blocks = defaultBlocks(renderCtx)
	}

	return renderCtx
}

func coalesceStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

var idMonthNames = map[time.Month]string{
	time.January: "Januari", time.February: "Februari", time.March: "Maret",
	time.April: "April", time.May: "Mei", time.June: "Juni",
	time.July: "Juli", time.August: "Agustus", time.September: "September",
	time.October: "Oktober", time.November: "November", time.December: "Desember",
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%d %s %d", t.Day(), idMonthNames[t.Month()], t.Year())
}

func buildFields(employeeFields []string, emp *entity.EmployeeRenderData) []FieldRow {
	fieldLabels := map[string]string{
		"name":            "Nama",
		"identity_number": "NIK",
		"birth_info":      "Tempat, Tgl Lahir",
		"education":       "Pendidikan",
		"gender":          "Jenis Kelamin",
		"religion":        "Agama",
		"address":         "Alamat",
		"phone":           "No. Telepon",
	}

	fieldValues := map[string]string{
		"name":            emp.Name,
		"identity_number": emp.IdentityNumber,
		"birth_info":      emp.BirthInfo,
		"education":       emp.Education,
		"gender":          emp.Gender,
		"religion":        emp.Religion,
		"address":         emp.Address,
		"phone":           emp.Phone,
	}

	var rows []FieldRow
	for _, f := range employeeFields {
		f = strings.TrimSpace(f)
		label, ok := fieldLabels[f]
		if !ok {
			continue
		}
		rows = append(rows, FieldRow{
			Label: label,
			Value: fieldValues[f],
		})
	}
	return rows
}

func defaultBlocks(ctx *RenderContext) []RenderBlock {
	return []RenderBlock{
		{Type: "document_header", Title: "SURAT PERJANJIAN KERJA"},
		{Type: "article", Label: "PASAL 1", Title: "Para Pihak"},
		{Type: "first_party"},
		{Type: "second_party", Fields: []FieldRow{
			{Label: "Nama", Value: ctx.Employee.Name},
			{Label: "NIK", Value: ctx.Employee.IdentityNumber},
			{Label: "Tempat, Tgl Lahir", Value: ctx.Employee.BirthInfo},
			{Label: "Alamat", Value: ctx.Employee.Address},
			{Label: "Pendidikan", Value: ctx.Employee.Education},
			{Label: "Jenis Kelamin", Value: ctx.Employee.Gender},
			{Label: "Agama", Value: ctx.Employee.Religion},
			{Label: "No. Telepon", Value: ctx.Employee.Phone},
		}},
		{Type: "article", Label: "PASAL 2", Title: "Masa Berlakunya Perjanjian"},
		{Type: "paragraph", Label: "AYAT 1", Content: fmt.Sprintf("Perjanjian ini berlaku mulai %s.", ctx.Contract.StartDate)},
	}
}
