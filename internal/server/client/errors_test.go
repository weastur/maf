package client

import (
	"errors"
	"testing"
)

func TestErrUnknownResponseFormat(t *testing.T) {
	t.Parallel()

	expected := "unknown response format"
	if ErrUnknownResponseFormat.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, ErrUnknownResponseFormat.Error())
	}
}

func TestStatusCodeError_Error(t *testing.T) {
	t.Parallel()

	err := &StatusCodeError{Code: 404}
	expected := "bad status code: 404"

	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestServerError_Error(t *testing.T) {
	t.Parallel()

	err := &ServerError{Details: "internal server error"}
	expected := "server error: internal server error"

	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestErrorsAreComparable(t *testing.T) {
	t.Parallel()

	if !errors.Is(ErrUnknownResponseFormat, ErrUnknownResponseFormat) {
		t.Errorf("expected ErrUnknownResponseFormat to be comparable to itself")
	}
}
