package data

import "errors"

var (
	ErrInstanceAlreadyExists = errors.New("instance already exists")
	ErrInstanceNotFound      = errors.New("instance not found")
	ErrInvalidInstance       = errors.New("invalid instance")
	ErrInvalidInstanceDir    = errors.New("invalid instance directory")
	ErrTempDirAlreadyExists  = errors.New("temp directory already exists")
	ErrTempDirDoesNotExist   = errors.New("temp directory does not exist")
	ErrTempIsNotDir          = errors.New("temp is not a directory")
)
