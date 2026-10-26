package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPayload struct {
	UserID int64
	RoleID *int64
}

type DeviceTokenPayload struct {
	DeviceID        string
	DeviceSessionID string
}

type Claims struct {
	UserID          int64  `json:"user_id,omitempty"`
	RoleID          *int64 `json:"role_id,omitempty"`
	PrincipalType   string `json:"principal_type,omitempty"`
	DeviceID        string `json:"device_id,omitempty"`
	DeviceSessionID string `json:"device_session_id,omitempty"`
	Type            string `json:"type,omitempty"`
	jwt.RegisteredClaims
}

type JWTConfig struct {
	Secret string

	AccessTTL  time.Duration
	RefreshTTL time.Duration
}
