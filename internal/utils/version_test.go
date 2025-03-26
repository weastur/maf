package utils

import "testing"

func TestAppVersion(t *testing.T) {
	expected := "v0.0.1-dev0"
	actual := AppVersion()

	if actual != expected {
		t.Errorf("AppVersion() = %v; want %v", actual, expected)
	}
}
