package rating

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid rating input")
	ErrSelectionNotFound = errors.New("rating selection not found")
	ErrSelectionLocked   = errors.New("rating selection is locked")
)
