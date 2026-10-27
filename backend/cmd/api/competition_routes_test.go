package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/competition"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type routeCompetitionRepository struct {
	item competition.Competition
}

func (r *routeCompetitionRepository) ListPublic(context.Context, int32, int32) ([]competition.Competition, error) {
	return []competition.Competition{}, nil
}
func (r *routeCompetitionRepository) ListCalendar(context.Context, competition.CalendarFilter) ([]competition.Competition, error) {
	return []competition.Competition{}, nil
}
func (r *routeCompetitionRepository) GetPublic(context.Context, int64) (competition.Competition, error) {
	return r.item, nil
}
func (r *routeCompetitionRepository) ValidateReferences(context.Context, competition.CompetitionReferences) error {
	return nil
}
func (r *routeCompetitionRepository) Create(context.Context, competition.CreateCompetition) (competition.Competition, error) {
	return r.item, nil
}
func (r *routeCompetitionRepository) GetByID(context.Context, int64) (competition.Competition, error) {
	return r.item, nil
}
func (r *routeCompetitionRepository) UpdateDraft(context.Context, competition.UpdateCompetition) (competition.Competition, error) {
	return r.item, nil
}

func routeCompetition() competition.Competition {
	return competition.Competition{
		ID: 7, Name: "Кубок", DisciplineID: 1, DisciplineName: "Программирование",
		CompetitionTypeID: 2, CompetitionTypeName: "Кубок", CompetitionLevelID: 3,
		Format: "online", Status: "published", StartAt: time.Now().Add(2 * time.Hour),
		EndAt: time.Now().Add(3 * time.Hour), RegistrationStartsAt: time.Now().Add(-time.Hour),
		RegistrationEndsAt: time.Now().Add(time.Hour), OnlineURL: stringAddress("https://example.org"),
	}
}

func stringAddress(value string) *string { return &value }

type denyingRBACRepository struct{}

func (denyingRBACRepository) HasPermission(context.Context, int64, string) (bool, error) {
	return false, nil
}
func (denyingRBACRepository) CreateRole(context.Context, rbac.CreateRoleRequest) (rbac.Role, error) {
	return rbac.Role{}, nil
}
func (denyingRBACRepository) AssignRole(context.Context, int64, int64) error { return nil }
func (denyingRBACRepository) ListRoles(context.Context) ([]rbac.Role, error) { return nil, nil }
func (denyingRBACRepository) ListPermissions(context.Context) ([]rbac.Permission, error) {
	return nil, nil
}
func (denyingRBACRepository) ListSelfAssignableRoles(context.Context) ([]rbac.Role, error) {
	return nil, nil
}

func TestRegisterCompetitionRoutes(t *testing.T) {
	permissions := map[string]string{}
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions[r.Method] = permission
			handler(w, r)
		})
	}
	mux := http.NewServeMux()
	handler := competition.NewHandler(competition.NewService(&routeCompetitionRepository{item: routeCompetition()}))
	registerCompetitionRoutes(mux, handler, protected)

	for _, tc := range []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/api/v1/calendar/competitions?from=2026-11-01&to=2026-12-01", want: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/competitions", want: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/competitions/7", want: http.StatusOK},
		{method: http.MethodPost, path: "/api/v1/competitions", want: http.StatusUnauthorized},
		{method: http.MethodPatch, path: "/api/v1/competitions/7", want: http.StatusUnauthorized},
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
		if response.Code != tc.want {
			t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, response.Code, tc.want)
		}
	}
	if permissions[http.MethodPost] != competition.PermissionCreate {
		t.Fatalf("POST permission = %q", permissions[http.MethodPost])
	}
	if permissions[http.MethodPatch] != competition.PermissionUpdate {
		t.Fatalf("PATCH permission = %q", permissions[http.MethodPatch])
	}
	if _, protected := permissions[http.MethodGet]; protected {
		t.Fatal("public GET routes must not use authorization or RBAC middleware")
	}
}

func TestCompetitionWriteRoutesReturnProtectionStatus(t *testing.T) {
	manager := token.New("competition-route-test-secret", time.Minute, time.Hour)
	middleware := rbac.NewMiddleware(manager, denyingRBACRepository{})
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return middleware.Authenticate(middleware.Require(permission)(handler))
	}
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 42})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	mux := http.NewServeMux()
	registerCompetitionRoutes(mux, competition.NewHandler(competition.NewService(&routeCompetitionRepository{})), protected)

	for _, method := range []string{http.MethodPost, http.MethodPatch} {
		path := "/api/v1/competitions"
		if method == http.MethodPatch {
			path += "/7"
		}
		unauthenticated := httptest.NewRecorder()
		mux.ServeHTTP(unauthenticated, httptest.NewRequest(method, path, nil))
		if unauthenticated.Code != http.StatusUnauthorized {
			t.Fatalf("%s unauthenticated status = %d", method, unauthenticated.Code)
		}

		request := httptest.NewRequest(method, path, nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		forbidden := httptest.NewRecorder()
		mux.ServeHTTP(forbidden, request)
		if forbidden.Code != http.StatusForbidden {
			t.Fatalf("%s forbidden status = %d", method, forbidden.Code)
		}
	}
}
