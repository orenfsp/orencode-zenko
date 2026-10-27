package errs

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("Conflict")
)