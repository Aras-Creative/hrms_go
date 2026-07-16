package entity

import (
	"fmt"
	"strings"
)

type Gender string

const (
	GenderFemale Gender = "female"
	GenderMale   Gender = "male"
	GenderOther  Gender = "other"
)

var validGenders = []Gender{
	GenderFemale, GenderMale, GenderOther,
}

func ParseGender(s string) (Gender, error) {
	g := Gender(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validGenders {
		if g == v {
			return g, nil
		}
	}
	return "", fmt.Errorf("invalid gender: %s (must be female, male, or other)", s)
}

func (g Gender) IsValid() bool {
	for _, v := range validGenders {
		if g == v {
			return true
		}
	}
	return false
}
