package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppVersion(t *testing.T) {
	t.Parallel()

	version = "v1.0.0" // Set the version to a known value for testing
	expected := "v1.0.0"
	actual := AppVersion()

	assert.Equal(t, expected, actual, "expected version to be %s, but got %s", expected, actual)
}
