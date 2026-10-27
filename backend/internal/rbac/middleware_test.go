package rbac

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

func TestPermissionMiddlewareAllowsAuthorizedUser(t *testing.T) {
	manager := token.New("secret", time.Minute, time.Hour)
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 15})
	if err != nil {
		t.Fatal(err)
	}
	middleware := NewMiddleware(manager, &repositoryStub{allowed: true})
	handler := middleware.Authenticate(middleware.Require(PermissionRolesCreate)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok || identity.UserID != 15 {
				t.Fatalf("identity = %#v, ok = %v", identity, ok)
			}
			w.WriteHeader(http.StatusNoContent)
		},
	)))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
}

func TestPermissionMiddlewareRejectsMissingPermission(t *testing.T) {
	manager := token.New("secret", time.Minute, time.Hour)
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 15})
	if err != nil {
		t.Fatal(err)
	}
	middleware := NewMiddleware(manager, &repositoryStub{allowed: false})
	handler := middleware.Authenticate(middleware.Require(PermissionRolesCreate)(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {
			t.Fatal("protected handler must not be called")
		},
	)))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/roles", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestAuthenticationMiddlewareRejectsMissingToken(t *testing.T) {
	middleware := NewMiddleware(token.New("secret", time.Minute, time.Hour), &repositoryStub{})
	handler := middleware.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not be called")
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}
