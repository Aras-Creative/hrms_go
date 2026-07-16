package entity

import (
	"fmt"
	"regexp"
	"strings"
)

type Phone struct {
	value string
}

var phoneClean = regexp.MustCompile(`[\s\-]+`)

func NewPhone(value string) (Phone, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return Phone{}, fmt.Errorf("phone is required")
	}
	cleaned := phoneClean.ReplaceAllString(v, "")
	if len(cleaned) < 8 || len(cleaned) > 15 {
		return Phone{}, fmt.Errorf("phone must be 8-15 characters")
	}
	if cleaned[0] != '+' && (cleaned[0] < '0' || cleaned[0] > '9') {
		return Phone{}, fmt.Errorf("phone must start with + or a digit")
	}
	for _, c := range cleaned[1:] {
		if c < '0' || c > '9' {
			return Phone{}, fmt.Errorf("phone contains invalid characters")
		}
	}
	return Phone{value: v}, nil
}

func PhoneFromDB(value string) Phone {
	return Phone{value: value}
}

func (p Phone) String() string {
	return p.value
}

func (p Phone) IsEmpty() bool {
	return p.value == ""
}
