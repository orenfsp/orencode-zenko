package errs

import "errors"

var (
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrAccessDenied            = errors.New("access denied")
	ErrEmailAlreadyExists      = errors.New("email is already registered")
	ErrConsentRequired         = errors.New("personal data consent is required")
	ErrParentalConsentRequired = errors.New("parental consent document is required for users under 18")
	ErrInvalidRegistrationData = errors.New("invalid registration data")
	ErrInvalidSession          = errors.New("invalid or expired session")
	ErrRoleNotFound            = errors.New("role not found")
	ErrTelegramUsernameInvalid = errors.New("invalid telegram username")
	ErrTelegramUsernameTaken   = errors.New("telegram username is already in use")
)
