package config

import "errors"

var errExampleValidator = errors.New("example validation error")

type Validator interface {
	Validate() error
}

type exampleValidator struct{}

func (v *exampleValidator) Validate() error {
	return errExampleValidator
}
