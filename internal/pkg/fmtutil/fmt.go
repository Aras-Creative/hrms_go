package fmtutil

import "fmt"

func FormatCurrency(amount int64) string {
	s := fmt.Sprintf("%d", amount)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func FormatMoney(cents int64, currency string) string {
	abs := cents
	negative := false
	if abs < 0 {
		negative = true
		abs = -abs
	}

	formatted := FormatCurrency(abs)
	if negative {
		formatted = "-" + formatted
	}

	switch currency {
	case "IDR":
		return "Rp " + formatted
	default:
		return currency + " " + formatted
	}
}

func FormatMoneyFloat(amount float64, currency string) string {
	return FormatMoney(int64(amount), currency)
}
