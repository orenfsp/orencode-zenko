package auth

import (
	"context"
	"github.com/google/uuid"
	"github.com/Mccalin/vera-backend/internal/session"
	"github.com/Mccalin/vera-backend/internal/user"
	"github.com/Mccalin/vera-backend/pkg/token"
)

type UserProvider interface {
	GetUserByUsername(ctx context.Context, username string) (user.User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (user.User, error)
	
}
type TokenManager interface {
	GenerateAccessToken(token.TokenPayload) (string, error)
	GenerateRefreshToken(token.TokenPayload) (string, error)
}

type SessionProvider interface {
	CreateSession(ctx context.Context, data session.Create)(session.Response, error)
	GetSessionByRefreshHash(ctx context.Context, token string)(session.Response, error)
	UpdateSessionRefreshHash(ctx context.Context, data session.Update, id uuid.UUID)(session.Response, error)
	DeleteExpiredSessions(ctx context.Context)(error)
	DeleteSessionByID(ctx context.Context, id uuid.UUID)(error)
}