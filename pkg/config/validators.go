package config

// var errExampleValidator = errors.New("example validation error")

type validator interface {
	Validate() error
}

type exampleValidator struct{}

func (v *exampleValidator) Validate() error {
	return nil
}
