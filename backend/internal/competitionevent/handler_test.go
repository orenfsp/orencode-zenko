package competitionevent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type handlerRBACRepository struct {
	allowed bool
}

func (r handlerRBACRepository) HasPermission(context.Context, int64, string) (bool, error) {
	return r.allowed, nil
}
func (handlerRBACRepository) CreateRole(context.Context, rbac.CreateRoleRequest) (rbac.Role, error) {
	return rbac.Role{}, nil
}
func (handlerRBACRepository) AssignRole(context.Context, int64, int64) error { return nil }
func (handlerRBACRepository) ListRoles(context.Context) ([]rbac.Role, error) { return nil, nil }
func (handlerRBACRepository) ListPermissions(context.Context) ([]rbac.Permission, error) {
	return nil, nil
}
func (handlerRBACRepository) ListSelfAssignableRoles(context.Context) ([]rbac.Role, error) {
	return nil, nil
}
func (handlerRBACRepository) ListRolePermissions(context.Context, int64) ([]string, error) {
	return nil, nil
}

func TestHandlerPublicListAndGet(t *testing.T) {
	event := testEvent()
	repo := &repositoryStub{
		listPublicFn: func(context.Context, int64) ([]Event, error) { return []Event{event}, nil },
		getPublicFn:  func(context.Context, int64, int64) (Event, error) { return event, nil },
	}
	handler := NewHandler(NewService(repo))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/competitions/{competitionID}/events", handler.ListPublic)
	mux.HandleFunc("GET /api/v1/competitions/{competitionID}/events/{eventID}", handler.GetPublic)
	for _, path := range []string{"/api/v1/competitions/7/events", "/api/v1/competitions/7/events/9"} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"protocol_url"`) {
			t.Fatalf("GET %s = %d %s", path, response.Code, response.Body.String())
		}
	}
}

func TestHandlerMapsNotFoundAndDatabaseErrors(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want int
	}{
		{name: "missing competition", err: ErrCompetitionMissing, want: http.StatusNotFound},
		{name: "hidden competition", err: ErrEventMissing, want: http.StatusNotFound},
		{name: "database", err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{listPublicFn: func(context.Context, int64) ([]Event, error) { return nil, tt.err }}
			handler := NewHandler(NewService(repo))
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("competitionID", "7")
			handler.ListPublic(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}

func TestHandlerProtectedMethodsRequireAuthentication(t *testing.T) {
	handler := NewHandler(NewService(&repositoryStub{}))
	for _, call := range []func(http.ResponseWriter, *http.Request){handler.ListAdmin, handler.GetAdmin, handler.Create, handler.Update, handler.Delete} {
		response := httptest.NewRecorder()
		call(response, httptest.NewRequest(http.MethodGet, "/", nil))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", response.Code)
		}
	}
}

func TestHandlerRejectsInvalidPathID(t *testing.T) {
	handler := NewHandler(NewService(&repositoryStub{}))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("competitionID", "bad")
	handler.ListPublic(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestHandlerCreateWithRBACAndStrictJSON(t *testing.T) {
	repo := &repositoryStub{createDraftFn: func(context.Context, int64, *int64) (Event, error) {
		return testEvent(), nil
	}}
	handler := NewHandler(NewService(repo))
	manager := token.New("competition-event-handler-test-secret", time.Minute, time.Hour)
	middleware := rbac.NewMiddleware(manager, handlerRBACRepository{allowed: true})
	protected := middleware.Authenticate(middleware.Require(PermissionCreate)(http.HandlerFunc(handler.Create)))
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 42})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	for _, tt := range []struct {
		name, body string
		want       int
	}{
		{name: "valid", body: `{"age_group_id":3}`, want: http.StatusCreated},
		{name: "discipline cannot be supplied", body: `{"discipline_id":2}`, want: http.StatusBadRequest},
		{name: "protocol cannot be supplied", body: `{"protocol_url":"https://example.org"}`, want: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			request.SetPathValue("competitionID", "7")
			request.Header.Set("Authorization", "Bearer "+accessToken)
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func TestHandlerPermissionDenied(t *testing.T) {
	manager := token.New("competition-event-handler-denied-secret", time.Minute, time.Hour)
	middleware := rbac.NewMiddleware(manager, handlerRBACRepository{})
	handler := NewHandler(NewService(&repositoryStub{}))
	protected := middleware.Authenticate(middleware.Require(PermissionDelete)(http.HandlerFunc(handler.Delete)))
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 42})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}
