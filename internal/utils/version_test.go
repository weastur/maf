package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppVersion(t *testing.T) {
	t.Parallel()

	expected := version
	actual := AppVersion()

	assert.Equal(t, expected, actual, "expected version to be %s, but got %s", expected, actual)
}
