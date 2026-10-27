package errs

import "errors"

var (
	ErrInvalidID         = errors.New("invalid municipality id")
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrNotFound          = errors.New("municipality not found")
	ErrInvalidCreate     = errors.New("invalid municipality data")
	ErrNameConflict      = errors.New("municipality name already exists")
)
