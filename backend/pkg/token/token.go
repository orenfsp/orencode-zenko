package token

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(secret string, accessTTL time.Duration, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) GenerateAccessToken(payload TokenPayload) (string, error) {
	return m.generateToken(Claims{
		UserID:        payload.UserID,
		RoleID:        payload.RoleID,
		PrincipalType: "user",
		Type:          "access",
	}, time.Now().Add(m.accessTTL))
}

func (m *Manager) GenerateRefreshToken(payload TokenPayload) (string, error) {
	return m.generateToken(Claims{
		UserID:        payload.UserID,
		RoleID:        payload.RoleID,
		PrincipalType: "user",
		Type:          "refresh",
	}, time.Now().Add(m.refreshTTL))
}

func (m *Manager) GenerateIntegrationAccessToken(payload IntegrationTokenPayload) (string, error) {
	if payload.ClientID <= 0 {
		return "", errors.New("integration client id is required")
	}
	return m.generateToken(Claims{
		IntegrationClientID: payload.ClientID,
		PrincipalType:       "integration",
		Type:                "access",
	}, time.Now().Add(m.accessTTL))
}

func (m *Manager) GenerateDeviceAccessToken(
	payload DeviceTokenPayload,
	sessionExpiresAt time.Time,
) (string, error) {
	expiresAt := time.Now().Add(m.accessTTL)
	if sessionExpiresAt.Before(expiresAt) {
		expiresAt = sessionExpiresAt
	}
	if !expiresAt.After(time.Now()) {
		return "", errors.New("device session is expired")
	}

	return m.generateToken(Claims{
		PrincipalType:   "device",
		DeviceID:        payload.DeviceID,
		DeviceSessionID: payload.DeviceSessionID,
		Type:            "access",
	}, expiresAt)
}

func (m *Manager) GenerateDeviceRefreshToken(
	payload DeviceTokenPayload,
	sessionExpiresAt time.Time,
) (string, error) {
	if !sessionExpiresAt.After(time.Now()) {
		return "", errors.New("device session is expired")
	}

	return m.generateToken(Claims{
		PrincipalType:   "device",
		DeviceID:        payload.DeviceID,
		DeviceSessionID: payload.DeviceSessionID,
		Type:            "refresh",
	}, sessionExpiresAt)
}

func (m *Manager) generateToken(claims Claims, expiresAt time.Time) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.NewString(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(m.secret))
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	var claims Claims

	parsed, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return &claims, nil
}

func HashRefresh(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (m *Manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

func (m *Manager) AccessTTL() time.Duration {
	return m.accessTTL
}
