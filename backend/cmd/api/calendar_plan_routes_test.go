package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/calendarplan"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type routeCalendarPlanRepository struct{}

func (routeCalendarPlanRepository) List(context.Context, calendarplan.ListFilter) ([]calendarplan.CalendarPlan, error) {
	return []calendarplan.CalendarPlan{}, nil
}
func (routeCalendarPlanRepository) Get(context.Context, int64) (calendarplan.CalendarPlan, error) {
	return calendarplan.CalendarPlan{ID: 1, Name: "План", Year: 2026}, nil
}
func (routeCalendarPlanRepository) Create(_ context.Context, plan calendarplan.CalendarPlan) (calendarplan.CalendarPlan, error) {
	plan.ID = 1
	return plan, nil
}
func (routeCalendarPlanRepository) Update(_ context.Context, plan calendarplan.UpdateCalendarPlan) (calendarplan.CalendarPlan, error) {
	return calendarplan.CalendarPlan{ID: plan.ID, Name: plan.Name, Year: plan.Year}, nil
}
func (routeCalendarPlanRepository) ListCompetitions(context.Context, int64, int32, int32) ([]calendarplan.Competition, error) {
	return []calendarplan.Competition{}, nil
}

func TestRegisterCalendarPlanRoutesUsesRequiredPermissions(t *testing.T) {
	permissions := map[string]string{}
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions[r.Method+" "+r.URL.Path] = permission
			handler(w, r)
		})
	}
	mux := http.NewServeMux()
	handler := calendarplan.NewHandler(calendarplan.NewService(routeCalendarPlanRepository{}))
	registerCalendarPlanRoutes(mux, handler, protected)

	cases := []struct {
		method, path, body, permission string
		want                           int
	}{
		{http.MethodGet, "/api/v1/admin/calendar-plans", "", calendarplan.PermissionRead, http.StatusOK},
		{http.MethodGet, "/api/v1/admin/calendar-plans/1", "", calendarplan.PermissionRead, http.StatusOK},
		{http.MethodGet, "/api/v1/admin/calendar-plans/1/competitions", "", calendarplan.PermissionRead, http.StatusOK},
		{http.MethodPost, "/api/v1/admin/calendar-plans", `{"name":"План","year":2026}`, calendarplan.PermissionCreate, http.StatusCreated},
		{http.MethodPatch, "/api/v1/admin/calendar-plans/1", `{"name":"Новый план"}`, calendarplan.PermissionUpdate, http.StatusOK},
	}
	for _, tc := range cases {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if response.Code != tc.want {
			t.Fatalf("%s %s status = %d, want %d: %s", tc.method, tc.path, response.Code, tc.want, response.Body.String())
		}
		if got := permissions[tc.method+" "+tc.path]; got != tc.permission {
			t.Fatalf("%s %s permission = %q, want %q", tc.method, tc.path, got, tc.permission)
		}
	}
}

func TestCalendarPlanRoutesRejectMissingTokenAndPermission(t *testing.T) {
	manager := token.New("calendar-plan-route-test-secret", time.Minute, time.Hour)
	middleware := rbac.NewMiddleware(manager, denyingRBACRepository{})
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return middleware.Authenticate(middleware.Require(permission)(handler))
	}
	mux := http.NewServeMux()
	registerCalendarPlanRoutes(mux, calendarplan.NewHandler(calendarplan.NewService(routeCalendarPlanRepository{})), protected)
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 42})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/admin/calendar-plans", ""},
		{http.MethodGet, "/api/v1/admin/calendar-plans/1", ""},
		{http.MethodGet, "/api/v1/admin/calendar-plans/1/competitions", ""},
		{http.MethodPost, "/api/v1/admin/calendar-plans", `{"name":"План","year":2026}`},
		{http.MethodPatch, "/api/v1/admin/calendar-plans/1", `{"year":2026}`},
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s without token status = %d", tc.method, tc.path, response.Code)
		}

		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Authorization", "Bearer "+accessToken)
		response = httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s %s without permission status = %d", tc.method, tc.path, response.Code)
		}
	}
}
