package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate validates a struct and returns a formatted error if validation fails.
// It properly handles both ValidationErrors and InvalidValidationError without panicking.
func Validate(v interface{}) error {
	if err := validate.Struct(v); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			var msgs []string
			for _, e := range validationErrs {
				msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field(), e.Tag()))
			}
			return fmt.Errorf("%s", strings.Join(msgs, "; "))
		}
		// Handle InvalidValidationError (nil pointer, non-struct, etc.)
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}
