package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PaySlipSource string

const (
	PaySlipSourceAuto   PaySlipSource = "auto"
	PaySlipSourceManual PaySlipSource = "manual"
)

var validPaySlipSources = []PaySlipSource{
	PaySlipSourceAuto,
	PaySlipSourceManual,
}

func ParsePaySlipSource(s string) (PaySlipSource, error) {
	src := PaySlipSource(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validPaySlipSources {
		if src == v {
			return src, nil
		}
	}
	return "", fmt.Errorf("invalid payslip source: %s (must be auto or manual)", s)
}

func (s PaySlipSource) IsValid() bool {
	for _, v := range validPaySlipSources {
		if s == v {
			return true
		}
	}
	return false
}

type CompensationBreakdown struct {
	CompensationItemID string  `json:"compensation_item_id"`
	Name               string  `json:"name"`
	Amount             float64 `json:"amount"`
}

type DeductionBreakdown struct {
	DeductionTypeID string  `json:"deduction_type_id"`
	Name            string  `json:"name"`
	Amount          float64 `json:"amount"`
}

type PaySlip struct {
	ID                     string
	PeriodID               string
	EmployeeID             string
	BaseSalary             Amount
	TotalCompensations     Amount
	TotalDeductions        Amount
	AbsentDays             int
	NetSalary              Amount
	Currency               Currency
	Source                 PaySlipSource
	CompensationsBreakdown []CompensationBreakdown
	DeductionsBreakdown    []DeductionBreakdown
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// --- PaySlipBuilder ---

type PaySlipBuilder struct {
	periodID   string
	employeeID string
	currency   Currency
	source     PaySlipSource
	baseSalary Amount
	absentDays int

	comps []CompensationBreakdown
	deds  []DeductionBreakdown
}

func NewPaySlipBuilder(periodID, employeeID string) *PaySlipBuilder {
	return &PaySlipBuilder{
		periodID:   periodID,
		employeeID: employeeID,
		currency:   CurrencyFromDB("IDR"),
		source:     PaySlipSourceAuto,
	}
}

func (b *PaySlipBuilder) WithCurrency(c Currency) *PaySlipBuilder {
	b.currency = c
	return b
}

func (b *PaySlipBuilder) WithSource(s PaySlipSource) *PaySlipBuilder {
	b.source = s
	return b
}

func (b *PaySlipBuilder) WithBaseSalary(a Amount) *PaySlipBuilder {
	b.baseSalary = a
	return b
}

func (b *PaySlipBuilder) WithAbsentDays(days int) *PaySlipBuilder {
	b.absentDays = days
	return b
}

func (b *PaySlipBuilder) AddCompensation(id, name string, amountCents int64) *PaySlipBuilder {
	b.comps = append(b.comps, CompensationBreakdown{
		CompensationItemID: id,
		Name:               name,
		Amount:             float64(amountCents) / 100,
	})
	return b
}

func (b *PaySlipBuilder) AddDeduction(id, name string, amountCents int64) *PaySlipBuilder {
	b.deds = append(b.deds, DeductionBreakdown{
		DeductionTypeID: id,
		Name:            name,
		Amount:          float64(amountCents) / 100,
	})
	return b
}

func (ps *PaySlip) CalculateNetSalary() Amount {
	totalCompCents := int64(0)
	for _, c := range ps.CompensationsBreakdown {
		totalCompCents += int64(c.Amount * 100)
	}
	totalDedCents := int64(0)
	for _, d := range ps.DeductionsBreakdown {
		totalDedCents += int64(d.Amount * 100)
	}
	netCents := ps.BaseSalary.Cents() + totalCompCents - totalDedCents
	return AmountFromCents(netCents)
}

func (b *PaySlipBuilder) Build() *PaySlip {
	var totalCompCents int64
	for _, c := range b.comps {
		totalCompCents += int64(c.Amount * 100)
	}

	var totalDedCents int64
	for _, d := range b.deds {
		totalDedCents += int64(d.Amount * 100)
	}

	netCents := b.baseSalary.Cents() + totalCompCents - totalDedCents

	now := time.Now()
	ps := &PaySlip{
		ID:                     uuid.New().String(),
		PeriodID:               b.periodID,
		EmployeeID:             b.employeeID,
		BaseSalary:             b.baseSalary,
		TotalCompensations:     AmountFromCents(totalCompCents),
		TotalDeductions:        AmountFromCents(totalDedCents),
		AbsentDays:             b.absentDays,
		NetSalary:              AmountFromCents(netCents),
		Currency:               b.currency,
		Source:                 b.source,
		CompensationsBreakdown: b.comps,
		DeductionsBreakdown:    b.deds,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	ps.NetSalary = ps.CalculateNetSalary()
	return ps
}

func ReconstitutePaySlip(
	id, periodID, employeeID string,
	baseSalaryCents, totalCompCents, totalDedCents int64,
	absentDays int,
	netSalaryCents int64,
	currency string,
	source string,
	compJSON, dedJSON []byte,
	createdAt, updatedAt time.Time,
) *PaySlip {
	var compB []CompensationBreakdown
	if len(compJSON) > 0 {
		json.Unmarshal(compJSON, &compB)
	}
	var dedB []DeductionBreakdown
	if len(dedJSON) > 0 {
		json.Unmarshal(dedJSON, &dedB)
	}
	return &PaySlip{
		ID:                     id,
		PeriodID:               periodID,
		EmployeeID:             employeeID,
		BaseSalary:             AmountFromCents(baseSalaryCents),
		TotalCompensations:     AmountFromCents(totalCompCents),
		TotalDeductions:        AmountFromCents(totalDedCents),
		AbsentDays:             absentDays,
		NetSalary:              AmountFromCents(netSalaryCents),
		Currency:               CurrencyFromDB(currency),
		Source:                 PaySlipSource(source),
		CompensationsBreakdown: compB,
		DeductionsBreakdown:    dedB,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
	}
}
