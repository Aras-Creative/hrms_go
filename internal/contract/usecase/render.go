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
}

//go:embed templates/contract.html
var contractTemplate string

//go:embed templates/company_logo.jpeg
var companyLogo []byte

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
	Employee    entity.EmployeeRenderData
	Contract    entity.ContractRenderData
	Signatory   entity.SignatoryRenderData
	Data        entity.ContractTemplateData
	Signings    []entity.ContractSigningRenderData
	Blocks      []RenderBlock
	CompanyLogo string
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
	pdf          PDFRenderer
}

func NewRenderUsecase(contractRepo repository.ContractRepository, signingRepo repository.SigningRepository, empFetcher EmployeeFetcher, pdf PDFRenderer) *RenderUsecase {
	return &RenderUsecase{contractRepo: contractRepo, signingRepo: signingRepo, empFetcher: empFetcher, pdf: pdf}
}

func (uc *RenderUsecase) Preview(ctx context.Context, contractID string, signatoryName, signatoryTitle string) ([]byte, error) {
	c, err := uc.contractRepo.FindContractByID(ctx, contractID)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find contract: %v", err))
	}
	if c == nil {
		return nil, errors.NewNotFound("contract not found")
	}

	emp, err := uc.empFetcher.FindEmployeeRenderData(ctx, c.EmployeeID)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find employee: %v", err))
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
		return nil, errors.NewInternal(fmt.Sprintf("failed to find contract: %v", err))
	}
	if c == nil {
		return nil, errors.NewNotFound("contract not found")
	}

	emp, err := uc.empFetcher.FindEmployeeRenderData(ctx, c.EmployeeID)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find employee: %v", err))
	}
	if emp == nil {
		return nil, errors.NewNotFound("employee not found")
	}

	return uc.renderAndGeneratePDF(ctx, c, emp, signatoryName, signatoryTitle, signings)
}

func (uc *RenderUsecase) renderAndGeneratePDF(ctx context.Context, c *entity.Contract, emp *entity.EmployeeRenderData, signatoryName, signatoryTitle string, signings []*entity.ContractSigning) ([]byte, error) {
	html, err := uc.renderHTML(c, emp, signatoryName, signatoryTitle, signings)
	if err != nil {
		return nil, err
	}

	return uc.pdf.Render(ctx, html)
}

func (uc *RenderUsecase) renderHTML(c *entity.Contract, emp *entity.EmployeeRenderData, signatoryName, signatoryTitle string, signings []*entity.ContractSigning) ([]byte, error) {
	renderCtx := buildRenderContext(c, emp, signatoryName, signatoryTitle, signings)

	// Resolve variable placeholders in block content
	resolveBlockContent(renderCtx.Blocks, renderCtx)

	tmpl, err := template.New("contract").Funcs(tmplFuncs).Parse(contractTemplate)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to parse template: %v", err))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderCtx); err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to execute template: %v", err))
	}

	return buf.Bytes(), nil
}

func buildRenderContext(c *entity.Contract, emp *entity.EmployeeRenderData, signatoryName, signatoryTitle string, signings []*entity.ContractSigning) *RenderContext {
	renderCtx := &RenderContext{
		Employee:    *emp,
		Contract:    buildContractRenderData(c),
		Signatory:   entity.SignatoryRenderData{Name: signatoryName, Designation: signatoryTitle},
		Data:        c.Data,
		Signings:    buildSigningRenderData(signings),
		CompanyLogo: base64.StdEncoding.EncodeToString(companyLogo),
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

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2 January 2006")
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
