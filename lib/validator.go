package lib

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		for _, tagName := range []string{"json", "form"} {
			name, _, _ := strings.Cut(field.Tag.Get(tagName), ",")
			if name != "" && name != "-" {
				return name
			}
		}

		return field.Name
	})

	return v
}

func ValidateStruct(value any) error {
	return validate.Struct(value)
}

func ParseFormErrors(err error) []FormError {
	var validateErrs validator.ValidationErrors
	if !errors.As(err, &validateErrs) {
		return []FormError{
			{
				Field:   "",
				Rule:    "invalid",
				Message: err.Error(),
			},
		}
	}

	parsedErrors := make([]FormError, 0, len(validateErrs))

	for _, validateErr := range validateErrs {
		parsedErrors = append(parsedErrors, FormError{
			Field:   validateErr.Field(),
			Rule:    validateErr.Tag(),
			Message: validationMessage(validateErr),
		})
	}

	return parsedErrors
}

func ParseValidationError(err error) []FormError {
	return ParseFormErrors(err)
}

func validationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}
