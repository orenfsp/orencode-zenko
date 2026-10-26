package auth

import (
	"context"
	"time"
)

type User struct {
	ID               int64
	Email            string
	PasswordHash     string
	TelegramUsername *string
	RoleID           *int64
	AccessStatus     bool
}

type CreateUser struct {
	FirstName                  string
	LastName                   string
	MiddleName                 *string
	BirthDate                  time.Time
	Email                      string
	PasswordHash               string
	ConsentPolicyVersion       string
	ParentalConsentDocumentURL *string
}

type Session struct {
	ID               int64
	UserID           int64
	RefreshTokenHash string
	DeviceID         string
	ExpiresAt        time.Time
}

type CreateSession struct {
	UserID           int64
	RefreshTokenHash string
	DeviceID         string
	ExpiresAt        time.Time
}

type UserProvider interface {
	CreateUser(ctx context.Context, data CreateUser) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id int64) (User, error)
	SetRole(ctx context.Context, userID, roleID int64) error
	UpdateTelegramUsername(ctx context.Context, userID int64, username *string) error
}

type SessionProvider interface {
	CreateSession(ctx context.Context, data CreateSession) (Session, error)
	GetSessionByRefreshHash(ctx context.Context, hash string) (Session, error)
	RotateSession(ctx context.Context, id int64, refreshHash string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, id int64) error
}
