package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/competitionevent"
)

type routeEventRepository struct{}

func (routeEventRepository) ListPublic(context.Context, int64) ([]competitionevent.Event, error) {
	return []competitionevent.Event{}, nil
}
func (routeEventRepository) GetPublic(context.Context, int64, int64) (competitionevent.Event, error) {
	return competitionevent.Event{}, competitionevent.ErrEventMissing
}
func (routeEventRepository) ListAdmin(context.Context, int64) ([]competitionevent.Event, error) {
	return []competitionevent.Event{}, nil
}
func (routeEventRepository) GetAdmin(context.Context, int64, int64) (competitionevent.Event, error) {
	return competitionevent.Event{}, competitionevent.ErrEventMissing
}
func (routeEventRepository) CreateDraft(context.Context, int64, *int64) (competitionevent.Event, error) {
	return competitionevent.Event{}, nil
}
func (routeEventRepository) UpdateDraft(context.Context, int64, int64, *int64) (competitionevent.Event, error) {
	return competitionevent.Event{}, nil
}
func (routeEventRepository) DeleteDraft(context.Context, int64, int64) error { return nil }

func TestRegisterCompetitionEventRoutes(t *testing.T) {
	permissions := map[string]string{}
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions[r.Method+" "+r.URL.Path] = permission
			handler(w, r)
		})
	}
	mux := http.NewServeMux()
	handler := competitionevent.NewHandler(competitionevent.NewService(routeEventRepository{}))
	registerCompetitionEventRoutes(mux, handler, protected)

	tests := []struct {
		method, path, permission string
		want                     int
	}{
		{method: http.MethodGet, path: "/api/v1/competitions/7/events", want: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/competitions/7/events/9", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/api/v1/admin/competitions/7/events", permission: competitionevent.PermissionRead, want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/admin/competitions/7/events/9", permission: competitionevent.PermissionRead, want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/competitions/7/events", permission: competitionevent.PermissionCreate, want: http.StatusUnauthorized},
		{method: http.MethodPatch, path: "/api/v1/competitions/7/events/9", permission: competitionevent.PermissionUpdate, want: http.StatusUnauthorized},
		{method: http.MethodDelete, path: "/api/v1/competitions/7/events/9", permission: competitionevent.PermissionDelete, want: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, nil))
		if response.Code != tt.want {
			t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, response.Code, tt.want)
		}
		if tt.permission != "" && permissions[tt.method+" "+tt.path] != tt.permission {
			t.Fatalf("%s %s permission = %q, want %q", tt.method, tt.path, permissions[tt.method+" "+tt.path], tt.permission)
		}
	}
}
