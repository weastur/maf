package config

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(viperInstance *viper.Viper) error {
	args := m.Called(viperInstance)

	return args.Error(0)
}

func TestGetSingletonInstance(t *testing.T) {
	t.Parallel()

	config1 := Get()
	config2 := Get()

	assert.Equal(t, config1, config2, "Get() should return the same instance")
}

func TestInitWithConfigFile(t *testing.T) {
	t.Parallel()

	mockValidator := new(MockValidator)
	mockValidator.On("Validate", mock.Anything).Return(nil)

	config := &Config{
		viperInstance: viper.New(),
		validators:    []Validator{mockValidator},
	}

	tempFile, err := os.CreateTemp(t.TempDir(), "test_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("key: value")
	require.NoError(t, err)
	tempFile.Close()

	err = config.Init(tempFile.Name())
	require.NoError(t, err, "Init should not return an error with a valid config file")

	mockValidator.AssertCalled(t, "Validate", mock.Anything)
}

func TestInitWithoutConfigFile(t *testing.T) {
	t.Parallel()

	mockValidator := new(MockValidator)
	mockValidator.On("Validate", mock.Anything).Return(nil)

	config := &Config{
		viperInstance: viper.New(),
		validators:    []Validator{mockValidator},
	}

	err := config.Init("")
	require.NoError(t, err, "Init should not return an error when no config file is provided")

	mockValidator.AssertCalled(t, "Validate", mock.Anything)
}

func TestInitWithInvalidConfigFile(t *testing.T) {
	t.Parallel()

	mockValidator := new(MockValidator)

	config := &Config{
		viperInstance: viper.New(),
		validators:    []Validator{mockValidator},
	}

	err := config.Init("nonexistent_file.yaml")
	require.Error(t, err, "Init should return an error for a nonexistent config file")
}

func TestValidateSuccess(t *testing.T) {
	t.Parallel()

	mockValidator1 := new(MockValidator)
	mockValidator2 := new(MockValidator)

	mockValidator1.On("Validate", mock.Anything).Return(nil)
	mockValidator2.On("Validate", mock.Anything).Return(nil)

	config := &Config{
		viperInstance: viper.New(),
		validators:    []Validator{mockValidator1, mockValidator2},
	}

	err := config.validate()
	require.NoError(t, err, "validate should not return an error when all validators pass")

	mockValidator1.AssertCalled(t, "Validate", mock.Anything)
	mockValidator2.AssertCalled(t, "Validate", mock.Anything)
}

func TestValidateFailure(t *testing.T) {
	t.Parallel()

	mockValidator1 := new(MockValidator)
	mockValidator2 := new(MockValidator)

	validationError := errors.New("validation error")

	mockValidator1.On("Validate", mock.Anything).Return(nil)
	mockValidator2.On("Validate", mock.Anything).Return(validationError)

	config := &Config{
		viperInstance: viper.New(),
		validators:    []Validator{mockValidator1, mockValidator2},
	}

	err := config.validate()
	require.ErrorIs(t, err, validationError, "validate should return the validation error from the second validator")

	mockValidator1.AssertCalled(t, "Validate", mock.Anything)
	mockValidator2.AssertCalled(t, "Validate", mock.Anything)
}

func TestViper(t *testing.T) {
	t.Parallel()

	config := &Config{
		viperInstance: viper.New(),
	}

	viperInstance := config.Viper()
	assert.Equal(t, config.viperInstance, viperInstance, "Viper() should return the same viper instance as in the config")
}
