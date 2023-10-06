package keys

import "errors"

var (
	ErrInvalidNumberOfArgs    = errors.New("invalid number of arguments")
	ErrEmptyKeyName           = errors.New("key cannot be empty")
	ErrKeyContainsWhitespaces = errors.New("key cannot contain spaces")
	ErrInvalidKeyType         = errors.New("invalid key type. key type must be either 'ecdsa' or 'bls'")
	ErrInvalidPassword        = errors.New("invalid password")
)
