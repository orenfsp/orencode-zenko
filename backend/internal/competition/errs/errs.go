package errs

import "errors"

var (
	ErrInvalidPagination     = errors.New("invalid pagination")
	ErrInvalidID             = errors.New("invalid competition id")
	ErrNotFound              = errors.New("competition not found")
	ErrInvalidCreate         = errors.New("invalid competition data")
	ErrInvalidReference      = errors.New("invalid competition reference")
	ErrNameConflict          = errors.New("competition name already exists")
	ErrForbidden             = errors.New("competition creation forbidden")
	ErrInvalidUpdate         = errors.New("invalid competition update")
	ErrNotDraft              = errors.New("competition is not a draft")
	ErrInvalidCalendarPeriod = errors.New("invalid calendar period")
	ErrInvalidCalendarFilter = errors.New("invalid calendar filter")
	ErrInvalidTransition     = errors.New("invalid competition status transition")
	ErrNotOwner              = errors.New("competition is not owned by user")
)
