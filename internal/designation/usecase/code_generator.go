package usecase

import (
	"context"
	"strings"
	"unicode"
)

func (uc *DesignationUsecase) AcronymFromName(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '-'
	})

	if len(words) == 0 {
		return ""
	}

	if len(words) == 1 {
		upper := strings.ToUpper(words[0])
		if len(upper) <= 4 {
			return upper
		}
		if len(upper) > 3 {
			return upper[:3]
		}
		return upper
	}

	var b strings.Builder
	for _, w := range words {
		for _, r := range w {
			if unicode.IsLetter(r) {
				b.WriteRune(unicode.ToUpper(r))
				break
			}
		}
	}

	return b.String()
}

func (uc *DesignationUsecase) GenerateUniqueCode(ctx context.Context, name string) (string, error) {
	code := uc.AcronymFromName(name)
	if code == "" {
		code = "DES"
	}

	existing, err := uc.repo.FindByCode(ctx, code)
	if err != nil {
		return "", err
	}

	for existing != nil {
		code = uc.padNextLetter(code, name)
		existing, err = uc.repo.FindByCode(ctx, code)
		if err != nil {
			return "", err
		}
	}

	return code, nil
}

func (uc *DesignationUsecase) padNextLetter(current string, name string) string {
	cleaned := strings.ToUpper(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		return -1
	}, name))

	for _, c := range cleaned {
		if !strings.ContainsRune(current, c) {
			return current + string(c)
		}
	}

	return current + "X"
}
