package cli

import "errors"

var (
	ErrInvalidURL           = errors.New("invalid URL")
	ErrOptionWithoutDefault = errors.New("option without default value")
	ErrInvalidNumberOfArgs  = errors.New("invalid number of arguments")
)
