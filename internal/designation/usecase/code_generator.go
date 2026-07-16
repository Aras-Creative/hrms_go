package usecase

import (
	"strings"
	"unicode"
)

func (uc *DesignationUsecase) AcronymFromName(name string) string {
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return ""
	}
	var b strings.Builder
	for _, f := range fields {
		for _, r := range f {
			if unicode.IsLetter(r) {
				b.WriteRune(unicode.ToUpper(r))
				break
			}
		}
	}
	result := b.String()
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}
