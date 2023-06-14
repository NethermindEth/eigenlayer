package data

import "errors"

var (
	ErrInstanceAlreadyExists = errors.New("instance already exists")
	ErrInvalidInstance       = errors.New("invalid instance")
	ErrInvalidInstanceDir    = errors.New("invalid instance directory")
)
