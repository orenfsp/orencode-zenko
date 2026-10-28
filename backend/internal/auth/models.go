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

type Profile struct {
	ID               int64
	FirstName        string
	LastName         string
	MiddleName       *string
	BirthDate        time.Time
	Email            string
	TelegramUsername *string
	AccessStatus     bool
	Municipality     *ProfileReference
	Organization     *ProfileReference
	Roles            []ProfileRole
}

type ProfileReference struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProfileRole struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
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
