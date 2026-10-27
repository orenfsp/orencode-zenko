package auth

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	autherrs "github.com/orenfsp/orencode-zenko.git/internal/auth/errs"
	"github.com/orenfsp/orencode-zenko.git/pkg/password"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

var telegramUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{5,32}$`)

type preparedRegistration struct {
	firstName  string
	lastName   string
	middleName *string
	birthDate  time.Time
	email      string
	minor      bool
}

type Service struct {
	users    UserProvider
	tokens   *token.Manager
	sessions SessionProvider
	now      func() time.Time
}

func NewService(users UserProvider, tokens *token.Manager, sessions SessionProvider) *Service {
	return &Service{
		users:    users,
		tokens:   tokens,
		sessions: sessions,
		now:      time.Now,
	}
}

func (s *Service) ValidateRegistration(req RegisterRequest) (bool, error) {
	prepared, err := s.prepareRegistration(req)
	if err != nil {
		return false, err
	}
	return prepared.minor, nil
}

func (s *Service) Register(
	ctx context.Context,
	req RegisterRequest,
	personalDataConsentFileID int64,
	parentalConsentFileID *int64,
) (TokenResponse, error) {
	prepared, err := s.prepareRegistration(req)
	if err != nil {
		return TokenResponse{}, err
	}
	if personalDataConsentFileID <= 0 {
		return TokenResponse{}, autherrs.ErrConsentRequired
	}
	if prepared.minor && (parentalConsentFileID == nil || *parentalConsentFileID <= 0) {
		return TokenResponse{}, autherrs.ErrParentalConsentRequired
	}
	if !prepared.minor && parentalConsentFileID != nil {
		return TokenResponse{}, autherrs.ErrInvalidDocument
	}

	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return TokenResponse{}, err
	}

	createdUser, err := s.users.CreateUser(ctx, CreateUser{
		FirstName:                 prepared.firstName,
		LastName:                  prepared.lastName,
		MiddleName:                prepared.middleName,
		BirthDate:                 prepared.birthDate,
		Email:                     prepared.email,
		PasswordHash:              passwordHash,
		PersonalDataConsentFileID: personalDataConsentFileID,
		ParentalConsentFileID:     parentalConsentFileID,
	})
	if err != nil {
		return TokenResponse{}, err
	}

	return s.createSession(ctx, createdUser)
}

func (s *Service) prepareRegistration(req RegisterRequest) (preparedRegistration, error) {
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	middleNameValue := strings.TrimSpace(req.MiddleName)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if firstName == "" || utf8.RuneCountInString(firstName) > 100 ||
		lastName == "" || utf8.RuneCountInString(lastName) > 100 ||
		utf8.RuneCountInString(middleNameValue) > 100 ||
		len(req.Password) < 8 || len(req.Password) > 72 ||
		len(email) > 255 || !validEmail(email) {
		return preparedRegistration{}, autherrs.ErrInvalidRegistrationData
	}
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return preparedRegistration{}, autherrs.ErrInvalidRegistrationData
	}
	today := dateOnly(s.now())
	if birthDate.After(today) || birthDate.Before(today.AddDate(-120, 0, 0)) {
		return preparedRegistration{}, autherrs.ErrInvalidRegistrationData
	}
	var middleName *string
	if middleNameValue != "" {
		middleName = &middleNameValue
	}
	return preparedRegistration{
		firstName:  firstName,
		lastName:   lastName,
		middleName: middleName,
		birthDate:  birthDate,
		email:      email,
		minor:      birthDate.AddDate(18, 0, 0).After(today),
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (TokenResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !validEmail(email) || req.Password == "" {
		return TokenResponse{}, autherrs.ErrInvalidCredentials
	}

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenResponse{}, autherrs.ErrInvalidCredentials
		}
		return TokenResponse{}, err
	}
	if err := password.Compare(req.Password, user.PasswordHash); err != nil {
		return TokenResponse{}, autherrs.ErrInvalidCredentials
	}
	if !user.AccessStatus {
		return TokenResponse{}, autherrs.ErrAccessDenied
	}

	return s.createSession(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, req RefreshTokenRequest) (TokenResponse, error) {
	claims, session, err := s.validRefreshSession(ctx, req.RefreshToken)
	if err != nil {
		return TokenResponse{}, err
	}

	user, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return TokenResponse{}, err
	}
	if !user.AccessStatus {
		return TokenResponse{}, autherrs.ErrAccessDenied
	}

	return s.rotateSession(ctx, session, user)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	_, session, err := s.validRefreshSession(ctx, refreshToken)
	if err != nil {
		return err
	}
	return s.sessions.DeleteSession(ctx, session.ID)
}

func (s *Service) PatchProfile(ctx context.Context, accessToken string, req PatchProfileRequest) (PatchProfileResponse, error) {
	claims, err := s.tokens.Parse(accessToken)
	if err != nil || claims.Type != "access" || claims.PrincipalType != "user" || claims.UserID == 0 {
		return PatchProfileResponse{}, autherrs.ErrAccessDenied
	}

	response := PatchProfileResponse{}
	if req.TelegramUsername != nil {
		value := strings.TrimSpace(*req.TelegramUsername)
		var username *string
		if value != "" {
			value = strings.TrimPrefix(value, "@")
			if !telegramUsernamePattern.MatchString(value) {
				return PatchProfileResponse{}, autherrs.ErrTelegramUsernameInvalid
			}
			username = &value
		}
		if err := s.users.UpdateTelegramUsername(ctx, claims.UserID, username); err != nil {
			return PatchProfileResponse{}, err
		}
		response.TelegramUsername = username
	}

	if req.RoleID == nil {
		return response, nil
	}
	if *req.RoleID <= 0 || strings.TrimSpace(req.RefreshToken) == "" {
		return PatchProfileResponse{}, autherrs.ErrInvalidRegistrationData
	}

	refreshClaims, session, err := s.validRefreshSession(ctx, req.RefreshToken)
	if err != nil || refreshClaims.UserID != claims.UserID {
		return PatchProfileResponse{}, autherrs.ErrInvalidSession
	}
	if err := s.users.SetRole(ctx, claims.UserID, *req.RoleID); err != nil {
		return PatchProfileResponse{}, err
	}
	user, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return PatchProfileResponse{}, err
	}
	tokens, err := s.rotateSession(ctx, session, user)
	if err != nil {
		return PatchProfileResponse{}, err
	}

	response.RoleID = user.RoleID
	response.AccessToken = tokens.AccessToken
	response.RefreshToken = tokens.RefreshToken
	return response, nil
}

func (s *Service) createSession(ctx context.Context, user User) (TokenResponse, error) {
	tokens, err := s.generateTokens(user)
	if err != nil {
		return TokenResponse{}, err
	}
	_, err = s.sessions.CreateSession(ctx, CreateSession{
		UserID:           user.ID,
		RefreshTokenHash: token.HashRefresh(tokens.RefreshToken),
		ExpiresAt:        s.now().Add(s.tokens.RefreshTTL()),
	})
	if err != nil {
		return TokenResponse{}, err
	}
	return tokens, nil
}

func (s *Service) rotateSession(ctx context.Context, session Session, user User) (TokenResponse, error) {
	tokens, err := s.generateTokens(user)
	if err != nil {
		return TokenResponse{}, err
	}
	err = s.sessions.RotateSession(
		ctx,
		session.ID,
		token.HashRefresh(tokens.RefreshToken),
		s.now().Add(s.tokens.RefreshTTL()),
	)
	if err != nil {
		return TokenResponse{}, err
	}
	return tokens, nil
}

func (s *Service) generateTokens(user User) (TokenResponse, error) {
	payload := token.TokenPayload{UserID: user.ID, RoleID: user.RoleID}
	accessToken, err := s.tokens.GenerateAccessToken(payload)
	if err != nil {
		return TokenResponse{}, err
	}
	refreshToken, err := s.tokens.GenerateRefreshToken(payload)
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) validRefreshSession(ctx context.Context, rawToken string) (*token.Claims, Session, error) {
	rawToken = strings.TrimSpace(rawToken)
	claims, err := s.tokens.Parse(rawToken)
	if err != nil || claims.Type != "refresh" || claims.PrincipalType != "user" || claims.UserID == 0 {
		return nil, Session{}, autherrs.ErrInvalidSession
	}
	session, err := s.sessions.GetSessionByRefreshHash(ctx, token.HashRefresh(rawToken))
	if err != nil || session.UserID != claims.UserID || !session.ExpiresAt.After(s.now()) {
		return nil, Session{}, autherrs.ErrInvalidSession
	}
	return claims, session, nil
}

func validEmail(value string) bool {
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
