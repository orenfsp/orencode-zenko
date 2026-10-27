package calendarplan

import "errors"

var (
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrInvalidYear       = errors.New("invalid calendar plan year")
	ErrInvalidID         = errors.New("invalid calendar plan id")
	ErrInvalidCreate     = errors.New("invalid calendar plan data")
	ErrInvalidUpdate     = errors.New("invalid calendar plan update")
	ErrNotFound          = errors.New("calendar plan not found")
)
