package client

import (
	"errors"
	"fmt"
)

var ErrUnknownResponseFormat = errors.New("unknown response format")

type StatusCodeError struct {
	Code int
}

type ServerError struct {
	Details string
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("bad status code: %d", e.Code)
}

func (e *ServerError) Error() string {
	return "server error: " + e.Details
}
