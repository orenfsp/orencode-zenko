package integration

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid integration request")
	ErrInvalidCredentials = errors.New("invalid integration credentials")
	ErrUnauthorized       = errors.New("integration authentication required")
	ErrForbidden          = errors.New("integration permission denied")
	ErrNotFound           = errors.New("integration resource not found")
	ErrConflict           = errors.New("integration resource conflict")
)
