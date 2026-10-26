package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Mccalin/vera-backend/internal/auth/errs"
	"github.com/Mccalin/vera-backend/internal/session"
	"github.com/Mccalin/vera-backend/pkg/logger"
	"github.com/Mccalin/vera-backend/pkg/password"
	"github.com/Mccalin/vera-backend/pkg/token"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Service struct {
	users   UserProvider
	token   *token.Manager
	session SessionProvider
}

// NewService accepts an optional session provider for backwards compatibility
// with callers that only exercise password/token generation (for example unit
// tests). Login, refresh and logout require a configured session provider.
func NewService(users UserProvider, token *token.Manager, sessions ...SessionProvider) *Service {
	var sessionProvider SessionProvider
	if len(sessions) > 0 {
		sessionProvider = sessions[0]
	}
	return &Service{
		users:   users,
		token:   token,
		session: sessionProvider,
	}
}

func (s *Service) Login(ctx context.Context, r LoginRequest) (LoginResponse, error) {
	user, err := s.users.GetUserByUsername(ctx, r.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResponse{}, errs.ErrInvalidCredentials
		}
		return LoginResponse{}, err
	}
	err = password.Compare(r.Password, user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResponse{}, errs.ErrInvalidPassword
		}
		return LoginResponse{}, err
	}
	if user.AccesStatus != true {
		return LoginResponse{}, errs.ErrAccessStatus
	}
	payload := token.TokenPayload{
		UserID: user.ID,
		RoleID: user.RoleID,
	}

	accessToken, err := s.token.GenerateAccessToken(payload)
	if err != nil {
		return LoginResponse{}, err
	}
	refreshToken, err := s.token.GenerateRefreshToken(payload)

	if err != nil {
		return LoginResponse{}, err
	}
	refreshHash := token.HashRefresh(refreshToken)
	_, err = s.session.CreateSession(ctx, session.Create{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		DeviceID:         r.DeviceID,
		ExpiresAt:        time.Now().Add(s.token.RefreshTTL()),
	})

	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) Refresh(ctx context.Context, dto RefreshTokenDTO) (RefreshResponseDTO, error) {
	// паршу токен
	claims, err := s.token.Parse(dto.RefreshToken)
	if err != nil {
		return RefreshResponseDTO{}, err
	}

	if claims.Type != "refresh" {
		return RefreshResponseDTO{}, errors.New("invalid token type")
	}
	// получаю рефреш, точнее его хеш
	hash := token.HashRefresh(dto.RefreshToken)
	userSessions, err := s.session.GetSessionByRefreshHash(ctx, hash)
	if err != nil {
		return RefreshResponseDTO{}, err
	}
	// не сдохла ли сессия
	if userSessions.ExpiresAt.Before(time.Now()) {
		return RefreshResponseDTO{}, errors.New("expired session")
	}
	// не попутал ли данные
	uid, err := uuid.Parse(userSessions.UserID)
	if claims.UserID != userSessions.UserID {
		return RefreshResponseDTO{}, errors.New("user not valid")
	}
	if err != nil {
		logger.Log.Error("parse uid fail auth", zap.Error(err))
		return RefreshResponseDTO{}, err
	}
	// по анологии с логином запоковываю новые токены послке успешного рефреша
	user, err := s.users.GetUserById(ctx, uid)
	if err != nil {
		return RefreshResponseDTO{}, err
	}
	if !user.AccesStatus {
		return RefreshResponseDTO{}, errs.ErrAccessStatus
	}
	payload := token.TokenPayload{
		UserID: user.ID,
		RoleID: user.RoleID,
	}
	accessToken, err := s.token.GenerateAccessToken(payload)
	if err != nil {
		return RefreshResponseDTO{}, err
	}
	refreshToken, err := s.token.GenerateRefreshToken(payload)

	if err != nil {
		return RefreshResponseDTO{}, err
	}

	newHash := token.HashRefresh(refreshToken)
	uidu, err := uuid.Parse(userSessions.ID)
	if err != nil {
		logger.Log.Error("parse uid fail auth", zap.Error(err))
		return RefreshResponseDTO{}, err
	}
	_, err = s.session.UpdateSessionRefreshHash(ctx, session.Update{
		RefreshTokenHash: newHash,
		ExpiresAt:        time.Now().Add(s.token.RefreshTTL()),
	}, uidu)
	if err != nil {
		return RefreshResponseDTO{}, err
	}

	return RefreshResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
func (s *Service) Logout(ctx context.Context, dto LogoutDTO) error {
	claims, err := s.token.Parse(dto.RefreshToken)
	if err != nil {
		return err
	}
	if claims.Type != "refresh" {
		return errors.New("user not valid")
	}
	hash := token.HashRefresh(dto.RefreshToken)
	userSession, err := s.session.GetSessionByRefreshHash(ctx, hash)
	if err != nil {
		return err
	}
	if claims.UserID != userSession.UserID {
		return errors.New("invalid session")
	}
	uid, err := uuid.Parse(userSession.ID)
	if err != nil {
		return err
	}
	err = s.session.DeleteSessionByID(ctx, uid)
	if err != nil {
		return err
	}

	return nil
}
