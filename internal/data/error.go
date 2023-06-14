package data

import "errors"

var (
	ErrInstanceAlreadyExists = errors.New("instance already exists")
	ErrInstanceNotFound      = errors.New("instance not found")
	ErrInvalidInstance       = errors.New("invalid instance")
	ErrInvalidInstanceDir    = errors.New("invalid instance directory")
)
