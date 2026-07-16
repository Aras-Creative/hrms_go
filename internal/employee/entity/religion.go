package entity

import (
	"fmt"
	"strings"
)

type Religion string

const (
	ReligionIslam     Religion = "islam"
	ReligionChristian Religion = "christian"
	ReligionCatholic  Religion = "catholic"
	ReligionHindu     Religion = "hindu"
	ReligionBuddhist  Religion = "buddhist"
	ReligionConfucian Religion = "confucian"
	ReligionOther     Religion = "other"
)

var validReligions = []Religion{
	ReligionIslam, ReligionChristian, ReligionCatholic,
	ReligionHindu, ReligionBuddhist, ReligionConfucian, ReligionOther,
}

func ParseReligion(s string) (Religion, error) {
	r := Religion(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validReligions {
		if r == v {
			return r, nil
		}
	}
	return "", fmt.Errorf("invalid religion: %s", s)
}

func (r Religion) IsValid() bool {
	for _, v := range validReligions {
		if r == v {
			return true
		}
	}
	return false
}
