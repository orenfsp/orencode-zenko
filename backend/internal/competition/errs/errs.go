package errs

import "errors"

var (
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrInvalidID         = errors.New("invalid competition id")
	ErrNotFound          = errors.New("competition not found")
)
