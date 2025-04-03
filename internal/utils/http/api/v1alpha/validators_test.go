package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestNewXValidator(t *testing.T) {
	t.Parallel()

	validator := NewXValidator()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.validator)
}

func TestXValidator_Validate_Success(t *testing.T) {
	t.Parallel()

	validator := NewXValidator()

	data := TestStruct{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	err := validator.Validate(data)
	assert.NoError(t, err)
}

func TestXValidator_Validate_ValidationErrors(t *testing.T) {
	t.Parallel()

	validator := NewXValidator()

	data := TestStruct{
		Name:  "",
		Email: "invalid-email",
	}

	err := validator.Validate(data)

	var validationError *ValidationError

	require.ErrorAs(t, err, &validationError)

	assert.Equal(t, "Name", validationError.Field)
	assert.Equal(t, "required", validationError.Tag)
	assert.Contains(t, err.Error(), "field 'Name' failed on tag 'required'")
}

func TestXValidator_Validate_InvalidValidationError(t *testing.T) {
	t.Parallel()

	validator := NewXValidator()

	err := validator.Validate(nil)
	require.Error(t, err)
}
