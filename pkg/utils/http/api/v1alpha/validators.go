package v1alpha

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type XValidator struct {
	validator *validator.Validate
}

type ValidationError struct {
	Field string
	Tag   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("field '%s' failed on tag '%s'", e.Field, e.Tag)
}

func NewXValidator() *XValidator {
	return &XValidator{
		validator: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (v *XValidator) Validate(data any) error {
	var validationErrors validator.ValidationErrors

	err := v.validator.Struct(data)

	switch {
	case errors.Is(err, &validator.InvalidValidationError{}):
		return fmt.Errorf("validation failed: %w", err)
	case errors.As(err, &validationErrors):
		errs := make([]error, 0, len(validationErrors))
		for _, err := range validationErrors {
			errs = append(errs, &ValidationError{
				Field: err.Field(),
				Tag:   err.Tag(),
			})
		}

		return errors.Join(errs...)
	case err != nil:
		return fmt.Errorf("unexpected error: %w", err)
	}

	return nil
}
