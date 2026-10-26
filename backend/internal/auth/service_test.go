package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	autherrs "github.com/Mccalin/vera-backend/internal/auth/errs"
	"github.com/Mccalin/vera-backend/internal/session"
	"github.com/Mccalin/vera-backend/internal/user"
	"github.com/Mccalin/vera-backend/pkg/password"
	"github.com/Mccalin/vera-backend/pkg/token"
	"github.com/google/uuid"
)

// Mock UserProvider
type mockUserProvider struct {
	user user.User
	err  error
}

func (m *mockUserProvider) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	if m.err != nil {
		return user.User{}, m.err
	}
	return m.user, nil
}

func (m *mockUserProvider) GetUserById(ctx context.Context, id uuid.UUID) (user.User, error) {
	if m.err != nil {
		return user.User{}, m.err
	}
	return m.user, nil
}

type mockSessionProvider struct{}

func (mockSessionProvider) CreateSession(context.Context, session.Create) (session.Response, error) {
	return session.Response{}, nil
}
func (mockSessionProvider) GetSessionByRefreshHash(context.Context, string) (session.Response, error) {
	return session.Response{}, errors.New("not implemented")
}
func (mockSessionProvider) UpdateSessionRefreshHash(context.Context, session.Update, uuid.UUID) (session.Response, error) {
	return session.Response{}, nil
}
func (mockSessionProvider) DeleteExpiredSessions(context.Context) error        { return nil }
func (mockSessionProvider) DeleteSessionByID(context.Context, uuid.UUID) error { return nil }

type refreshSessionProvider struct {
	session session.Response
}

func (s refreshSessionProvider) CreateSession(context.Context, session.Create) (session.Response, error) {
	return session.Response{}, nil
}
func (s refreshSessionProvider) GetSessionByRefreshHash(context.Context, string) (session.Response, error) {
	return s.session, nil
}
func (s refreshSessionProvider) UpdateSessionRefreshHash(context.Context, session.Update, uuid.UUID) (session.Response, error) {
	return session.Response{}, nil
}
func (refreshSessionProvider) DeleteExpiredSessions(context.Context) error { return nil }
func (refreshSessionProvider) DeleteSessionByID(context.Context, uuid.UUID) error {
	return nil
}

// Tests
func TestLoginSuccess(t *testing.T) {
	ctx := context.Background()

	// Создаем password hash
	hashPassword, _ := password.Hash("testPassword123")

	// Mock user
	mockUser := user.User{
		ID:           "user-id-123",
		Username:     "testuser",
		RoleID:       "role-id-123",
		AccesStatus:  true,
		PasswordHash: hashPassword,
	}

	mockUserProvider := &mockUserProvider{
		user: mockUser,
	}

	// Создаем token manager
	tokenManager := token.New(
		"test-secret-key",
		5*time.Minute,
		24*time.Hour,
	)
	//session

	service := NewService(mockUserProvider, tokenManager, mockSessionProvider{})

	// Login request
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "testPassword123",
	}

	// Запускаем Login
	response, err := service.Login(ctx, loginReq)

	// Проверяем результаты
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.AccessToken == "" {
		t.Error("Expected access token, got empty string")
	}

	if response.RefreshToken == "" {
		t.Error("Expected refresh token, got empty string")
	}

	t.Log("✓ Login successful with valid credentials")
}

func TestLoginUserNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserProvider := &mockUserProvider{
		err: errors.New("user not found"),
	}

	tokenManager := token.New("test-secret-key", 5*time.Minute, 24*time.Hour)
	service := NewService(mockUserProvider, tokenManager, mockSessionProvider{})

	loginReq := LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	_, err := service.Login(ctx, loginReq)

	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}

	t.Log("✓ Login failed with non-existent user")
}

func TestRefreshRejectsDisabledUser(t *testing.T) {
	userID := uuid.NewString()
	manager := token.New("test-secret-key", 5*time.Minute, 24*time.Hour)
	refreshToken, err := manager.GenerateRefreshToken(token.TokenPayload{
		UserID: userID,
		RoleID: uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	service := NewService(
		&mockUserProvider{user: user.User{
			ID:          userID,
			RoleID:      uuid.NewString(),
			AccesStatus: false,
		}},
		manager,
		refreshSessionProvider{session: session.Response{
			ID:        uuid.NewString(),
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
		}},
	)

	_, err = service.Refresh(context.Background(), RefreshTokenDTO{RefreshToken: refreshToken})
	if !errors.Is(err, autherrs.ErrAccessStatus) {
		t.Fatalf("Refresh() error = %v, want ErrAccessStatus", err)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	ctx := context.Background()

	hashPassword, _ := password.Hash("correctPassword123")

	mockUser := user.User{
		ID:           "user-id-123",
		Username:     "testuser",
		RoleID:       "role-id-123",
		PasswordHash: hashPassword,
	}

	mockUserProvider := &mockUserProvider{
		user: mockUser,
	}

	tokenManager := token.New("test-secret-key", 5*time.Minute, 24*time.Hour)
	service := NewService(mockUserProvider, tokenManager, mockSessionProvider{})

	loginReq := LoginRequest{
		Username: "testuser",
		Password: "wrongPassword",
	}

	_, err := service.Login(ctx, loginReq)

	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}

	t.Log("✓ Login failed with wrong password")
}
