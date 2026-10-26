package errs

import ("errors")

var (
	ErrInvalidCredentials = errors.New("user not found")
	ErrInvalidPassword = errors.New("password not correct")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrPermissionIDNotFound = errors.New("id not found")
	InternalServerError = errors.New("internal server error")
	ErrNmaeCannotBeEmpty = errors.New("name can't be empty")
	ErrAccessStatus = errors.New("access denied")
)