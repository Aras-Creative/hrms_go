package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	return &Validator{
		v: validator.New(),
	}
}

func (v *Validator) Struct(s interface{}) error {
	return v.v.Struct(s)
}

func (v *Validator) Errors(err error) string {
	if err == nil {
		return ""
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		var msgs []string
		for _, e := range ve {
			msg := formatValidationError(e)
			msgs = append(msgs, msg)
		}
		return strings.Join(msgs, "; ")
	}

	return err.Error()
}

func formatValidationError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "min_username":
		return fmt.Sprintf("%s must be at least %s characters and alphanumeric", field, param)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

func (v *Validator) ErrorsMap(err error) map[string]string {
	if err == nil {
		return nil
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range ve {
			errors[e.Field()] = formatValidationError(e)
		}
		return errors
	}

	return map[string]string{
		"error": err.Error(),
	}
}

func (v *Validator) GetValidator() *validator.Validate {
	return v.v
}
