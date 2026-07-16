package entity

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

type Amount struct {
	cents int64
}

func NewAmount(value float64) (Amount, error) {
	if value < 0 {
		return Amount{}, fmt.Errorf("amount must be non-negative")
	}
	return Amount{cents: int64(math.Round(value * 100))}, nil
}

func AmountFromCents(cents int64) Amount {
	return Amount{cents: cents}
}

func (a Amount) Cents() int64 {
	return a.cents
}

func (a Amount) Float() float64 {
	return float64(a.cents) / 100
}

func (a Amount) IsZero() bool {
	return a.cents == 0
}

var currencyRegex = regexp.MustCompile(`^[A-Z]{3}$`)

type Currency struct {
	code string
}

func NewCurrency(code string) (Currency, error) {
	c := strings.ToUpper(strings.TrimSpace(code))
	if !currencyRegex.MatchString(c) {
		return Currency{}, fmt.Errorf("currency must be a 3-letter ISO code")
	}
	return Currency{code: c}, nil
}

func CurrencyFromDB(code string) Currency {
	return Currency{code: code}
}

func (c Currency) String() string {
	return c.code
}

type Percentage struct {
	value float64
}

func NewPercentage(value float64) (Percentage, error) {
	if value < 0 || value > 100 {
		return Percentage{}, fmt.Errorf("percentage must be between 0 and 100")
	}
	return Percentage{value: value}, nil
}

func PercentageFromDB(value float64) Percentage {
	return Percentage{value: value}
}

func (p Percentage) Value() float64 {
	return p.value
}

type ContributionType string

const (
	ContributionTypePercentage ContributionType = "percentage"
	ContributionTypeFixed      ContributionType = "fixed"
)

var validContributionTypes = []ContributionType{
	ContributionTypePercentage,
	ContributionTypeFixed,
}

func ParseContributionType(s string) (ContributionType, error) {
	ct := ContributionType(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validContributionTypes {
		if ct == v {
			return ct, nil
		}
	}
	return "", fmt.Errorf("invalid contribution type: %s (must be percentage or fixed)", s)
}

func (ct ContributionType) IsValid() bool {
	for _, v := range validContributionTypes {
		if ct == v {
			return true
		}
	}
	return false
}

type DeductionCalcType string

const (
	DeductionCalcPercentage DeductionCalcType = "percentage"
	DeductionCalcFixed      DeductionCalcType = "fixed"
)

var validDeductionCalcTypes = []DeductionCalcType{
	DeductionCalcPercentage,
	DeductionCalcFixed,
}

func ParseDeductionCalcType(s string) (DeductionCalcType, error) {
	dct := DeductionCalcType(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validDeductionCalcTypes {
		if dct == v {
			return dct, nil
		}
	}
	return "", fmt.Errorf("invalid deduction type: %s (must be percentage or fixed)", s)
}

func (dct DeductionCalcType) IsValid() bool {
	for _, v := range validDeductionCalcTypes {
		if dct == v {
			return true
		}
	}
	return false
}

type Frequency string

const (
	FrequencyMonthly Frequency = "monthly"
	FrequencyYearly  Frequency = "yearly"
	FrequencyOneTime Frequency = "one_time"
)

var validFrequencies = []Frequency{
	FrequencyMonthly,
	FrequencyYearly,
	FrequencyOneTime,
}

func ParseFrequency(s string) (Frequency, error) {
	f := Frequency(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validFrequencies {
		if f == v {
			return f, nil
		}
	}
	return "", fmt.Errorf("invalid frequency: %s (must be monthly, yearly, or one_time)", s)
}

func (f Frequency) IsValid() bool {
	for _, v := range validFrequencies {
		if f == v {
			return true
		}
	}
	return false
}

type CompensationItemType string

const (
	CompItemTypeRecurring CompensationItemType = "recurring"
	CompItemTypeOneTime   CompensationItemType = "one_time"
)

var validCompItemTypes = []CompensationItemType{
	CompItemTypeRecurring,
	CompItemTypeOneTime,
}

func ParseCompensationItemType(s string) (CompensationItemType, error) {
	ct := CompensationItemType(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validCompItemTypes {
		if ct == v {
			return ct, nil
		}
	}
	return "", fmt.Errorf("invalid compensation item type: %s (must be recurring or one_time)", s)
}

func (ct CompensationItemType) IsValid() bool {
	for _, v := range validCompItemTypes {
		if ct == v {
			return true
		}
	}
	return false
}
