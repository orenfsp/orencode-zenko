package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/internal/result"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type routeResultRepository struct{}

func (routeResultRepository) GetEventContext(context.Context, int64, int64) (result.EventContext, error) {
	return result.EventContext{CompetitionID: 1, EventID: 2, DisciplineName: result.AlgorithmicDiscipline}, nil
}
func (routeResultRepository) ListApprovedApplications(context.Context, int64) ([]result.Application, error) {
	return []result.Application{}, nil
}
func (routeResultRepository) ListEligibleTeamMemberships(context.Context, int64) ([]result.TeamMembership, error) {
	return []result.TeamMembership{}, nil
}
func (routeResultRepository) Configure(context.Context, int64, string, int16, []result.TeamSetup, bool) error {
	return nil
}
func (routeResultRepository) GetTable(context.Context, int64) (result.Table, error) {
	key := "protocols/test.pdf"
	approved := time.Now()
	return result.Table{Sheet: result.Sheet{
		ID: 1, EventID: 2, Mode: result.ModeIndividual, TaskCount: 1,
		Status: result.SheetApproved, ApprovedAt: &approved, ProtocolStorageKey: &key,
	}, Entries: []result.Entry{}}, nil
}
func (routeResultRepository) SaveDraftEntries(context.Context, int64, []result.EntryUpdate) error {
	return nil
}
func (routeResultRepository) SaveCalculated(context.Context, int64, int64, []result.EntrySummary) error {
	return nil
}
func (routeResultRepository) Approve(context.Context, int64, int64, int64, []result.EntrySummary, string, string) (result.Approval, error) {
	return result.Approval{}, nil
}
func (routeResultRepository) ListProtocolJudges(context.Context, int64) ([]result.Judge, error) {
	return []result.Judge{}, nil
}
func (routeResultRepository) ListHistory(context.Context, int64) ([]result.HistoryItem, error) {
	return nil, nil
}

type routeResultStorage struct{}

func (routeResultStorage) Put(context.Context, string, io.Reader) (int64, error) { return 0, nil }
func (routeResultStorage) Open(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte("%PDF-test"))), nil
}
func (routeResultStorage) Delete(context.Context, string) error { return nil }

type routeProtocolGenerator struct{}

func (routeProtocolGenerator) Generate(result.ProtocolData) ([]byte, error) {
	return []byte("%PDF-test"), nil
}

func resultRouteHandler() *result.Handler {
	return result.NewHandler(result.NewService(routeResultRepository{}, routeResultStorage{}, routeProtocolGenerator{}))
}

func TestCompetitionResultRoutesAndPermissions(t *testing.T) {
	permissions := make(map[string]string)
	protected := func(permission string, _ http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions[r.Method+" "+r.URL.Path] = permission
			w.WriteHeader(http.StatusNoContent)
		})
	}
	mux := http.NewServeMux()
	registerCompetitionResultRoutes(mux, resultRouteHandler(), protected)

	tests := []struct{ method, path, permission string }{
		{http.MethodGet, "/api/v1/admin/competitions/1/events/2/results", result.PermissionRead},
		{http.MethodPut, "/api/v1/admin/competitions/1/events/2/results/settings", result.PermissionManage},
		{http.MethodPut, "/api/v1/admin/competitions/1/events/2/results", result.PermissionManage},
		{http.MethodPost, "/api/v1/admin/competitions/1/events/2/results/calculate", result.PermissionManage},
		{http.MethodPost, "/api/v1/admin/competitions/1/events/2/results/approve", result.PermissionApprove},
	}
	for _, tc := range tests {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
		if response.Code != http.StatusNoContent || permissions[tc.method+" "+tc.path] != tc.permission {
			t.Fatalf("%s %s status=%d permission=%q", tc.method, tc.path, response.Code, permissions[tc.method+" "+tc.path])
		}
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/competitions/1/events/2/protocol", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("public protocol status=%d content-type=%q", response.Code, response.Header().Get("Content-Type"))
	}
}

func TestCompetitionResultAdminRoutesRequireTokenAndPermission(t *testing.T) {
	manager := token.New("competition-result-route-secret", time.Minute, time.Hour)
	middleware := rbac.NewMiddleware(manager, denyingRBACRepository{})
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return middleware.Authenticate(middleware.Require(permission)(handler))
	}
	mux := http.NewServeMux()
	registerCompetitionResultRoutes(mux, resultRouteHandler(), protected)
	path := "/api/v1/admin/competitions/1/events/2/results"

	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("without token status=%d", response.Code)
	}
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 42})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error=%v", err)
	}
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("without permission status=%d", response.Code)
	}
}
