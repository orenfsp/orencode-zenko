package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	autherrs "github.com/orenfsp/orencode-zenko.git/internal/auth/errs"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type userProviderStub struct {
	user       User
	createData CreateUser
	createErr  error
}

func (s *userProviderStub) CreateUser(_ context.Context, data CreateUser) (User, error) {
	s.createData = data
	return s.user, s.createErr
}

func (s *userProviderStub) GetUserByEmail(context.Context, string) (User, error) {
	return s.user, nil
}

func (s *userProviderStub) GetUserByID(context.Context, int64) (User, error) {
	return s.user, nil
}

func (s *userProviderStub) SetRole(_ context.Context, _ int64, roleID int64) error {
	s.user.RoleID = &roleID
	return nil
}

func (s *userProviderStub) UpdateTelegramUsername(_ context.Context, _ int64, username *string) error {
	s.user.TelegramUsername = username
	return nil
}

type sessionProviderStub struct {
	session    Session
	createData CreateSession
	rotated    bool
}

func (s *sessionProviderStub) CreateSession(_ context.Context, data CreateSession) (Session, error) {
	s.createData = data
	s.session = Session{
		ID:               10,
		UserID:           data.UserID,
		RefreshTokenHash: data.RefreshTokenHash,
		ExpiresAt:        data.ExpiresAt,
	}
	return s.session, nil
}

func (s *sessionProviderStub) GetSessionByRefreshHash(context.Context, string) (Session, error) {
	return s.session, nil
}

func (s *sessionProviderStub) RotateSession(_ context.Context, _ int64, hash string, expiresAt time.Time) error {
	s.rotated = true
	s.session.RefreshTokenHash = hash
	s.session.ExpiresAt = expiresAt
	return nil
}

func (s *sessionProviderStub) DeleteSession(context.Context, int64) error { return nil }

func newTestService(users *userProviderStub, sessions *sessionProviderStub) *Service {
	manager := token.New("test-secret", 5*time.Minute, 24*time.Hour)
	service := NewService(users, manager, sessions)
	service.now = func() time.Time {
		return time.Date(2026, time.October, 26, 12, 0, 0, 0, time.UTC)
	}
	return service
}

func validRegisterRequest() RegisterRequest {
	return RegisterRequest{
		FirstName: "Иван",
		LastName:  "Иванов",
		BirthDate: "2000-05-10",
		Email:     "USER@example.com",
		Password:  "password123",
	}
}

func TestRegisterAdultCreatesRolelessTokens(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42, Email: "user@example.com", AccessStatus: true}}
	sessions := &sessionProviderStub{}
	service := newTestService(users, sessions)

	response, err := service.Register(context.Background(), validRegisterRequest(), 101, nil)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if users.createData.Email != "user@example.com" {
		t.Fatalf("normalized email = %q", users.createData.Email)
	}
	if users.createData.PersonalDataConsentFileID != 101 {
		t.Fatalf("personal consent file id = %d", users.createData.PersonalDataConsentFileID)
	}
	if users.createData.ParentalConsentFileID != nil {
		t.Fatal("adult registration unexpectedly created parental consent")
	}
	claims, err := service.tokens.Parse(response.AccessToken)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != 42 || claims.RoleID != nil {
		t.Fatalf("claims = user:%d role:%v, want user:42 and no role", claims.UserID, claims.RoleID)
	}
}

func TestRegisterMinorRequiresParentalConsent(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42}}
	service := newTestService(users, &sessionProviderStub{})
	req := validRegisterRequest()
	req.BirthDate = "2008-10-27"

	_, err := service.Register(context.Background(), req, 101, nil)
	if !errors.Is(err, autherrs.ErrParentalConsentRequired) {
		t.Fatalf("Register() error = %v, want ErrParentalConsentRequired", err)
	}
}

func TestRegisterRequiresPersonalDataConsentFile(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42}}
	service := newTestService(users, &sessionProviderStub{})

	_, err := service.Register(context.Background(), validRegisterRequest(), 0, nil)
	if !errors.Is(err, autherrs.ErrConsentRequired) {
		t.Fatalf("Register() error = %v, want ErrConsentRequired", err)
	}
}

func TestRegisterMinorPersistsParentalDocument(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42}}
	service := newTestService(users, &sessionProviderStub{})
	req := validRegisterRequest()
	req.BirthDate = "2010-01-01"
	parentalFileID := int64(202)

	if _, err := service.Register(context.Background(), req, 101, &parentalFileID); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if users.createData.ParentalConsentFileID == nil || *users.createData.ParentalConsentFileID != parentalFileID {
		t.Fatalf("parental file was not persisted: %#v", users.createData.ParentalConsentFileID)
	}
}

func TestRegisterAdultRejectsParentalDocument(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42}}
	service := newTestService(users, &sessionProviderStub{})
	parentalFileID := int64(202)

	_, err := service.Register(context.Background(), validRegisterRequest(), 101, &parentalFileID)
	if !errors.Is(err, autherrs.ErrInvalidDocument) {
		t.Fatalf("Register() error = %v, want ErrInvalidDocument", err)
	}
}

func TestRegisterUserTurning18TodayDoesNotRequireParentalConsent(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42}}
	service := newTestService(users, &sessionProviderStub{})
	req := validRegisterRequest()
	req.BirthDate = "2008-10-26"

	if _, err := service.Register(context.Background(), req, 101, nil); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

func TestPatchProfileRoleRotatesTokensWithRole(t *testing.T) {
	users := &userProviderStub{user: User{ID: 42, Email: "user@example.com", AccessStatus: true}}
	sessions := &sessionProviderStub{}
	service := newTestService(users, sessions)
	initialTokens, err := service.createSession(context.Background(), users.user)
	if err != nil {
		t.Fatalf("createSession() error = %v", err)
	}
	roleID := int64(7)

	response, err := service.PatchProfile(context.Background(), initialTokens.AccessToken, PatchProfileRequest{
		RoleID:       &roleID,
		RefreshToken: initialTokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("PatchProfile() error = %v", err)
	}
	if !sessions.rotated {
		t.Fatal("refresh session was not rotated")
	}
	claims, err := service.tokens.Parse(response.AccessToken)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.RoleID == nil || *claims.RoleID != roleID {
		t.Fatalf("role claim = %v, want %d", claims.RoleID, roleID)
	}
}
