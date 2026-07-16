package usecase

import (
	"regexp"
	"strings"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func generateSlug(name string) string {
	s := strings.TrimSpace(name)
	s = nonAlphanumeric.ReplaceAllString(s, "_")
	s = strings.ToLower(s)
	s = strings.Trim(s, "_")
	if len(s) > 50 {
		s = s[:50]
		s = strings.TrimRight(s, "_")
	}
	return s
}
