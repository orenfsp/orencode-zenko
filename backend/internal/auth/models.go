package auth

import "time"

type User struct {
	ID               int64
	Email            string
	PasswordHash     string
	TelegramUsername *string
	RoleID           *int64
	AccessStatus     bool
}

type CreateUser struct {
	FirstName                 string
	LastName                  string
	MiddleName                *string
	BirthDate                 time.Time
	Email                     string
	PasswordHash              string
	PersonalDataConsentFileID int64
	ParentalConsentFileID     *int64
}

type Session struct {
	ID               int64
	UserID           int64
	RefreshTokenHash string
	ExpiresAt        time.Time
}

type CreateSession struct {
	UserID           int64
	RefreshTokenHash string
	ExpiresAt        time.Time
}
