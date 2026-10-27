package rbac

import "errors"

var (
	ErrUnauthenticated    = errors.New("authentication required")
	ErrForbidden          = errors.New("permission denied")
	ErrInvalidRole        = errors.New("invalid role")
	ErrRoleAlreadyExists  = errors.New("role already exists")
	ErrRoleNotFound       = errors.New("role not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrPermissionNotFound = errors.New("permission not found")
)
