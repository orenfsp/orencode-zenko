package errs

import "errors"

var (
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrInvalidID         = errors.New("invalid competition id")
	ErrNotFound          = errors.New("competition not found")
	ErrInvalidCreate     = errors.New("invalid competition data")
	ErrInvalidReference  = errors.New("invalid competition reference")
	ErrNameConflict      = errors.New("competition name already exists")
	ErrForbidden         = errors.New("competition creation forbidden")
)
