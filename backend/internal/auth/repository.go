package auth

import (
	"context"
	"time"
)

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

type Repository interface {
	UserProvider
	SessionProvider
}
