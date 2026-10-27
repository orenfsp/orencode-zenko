package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/competition"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type routeJudgeRepository struct{}

func (routeJudgeRepository) GetCompetitionStatus(context.Context, int64) (string, error) {
	return "published", nil
}
func (routeJudgeRepository) ListCompetitionJudges(context.Context, int64) ([]competition.CompetitionJudge, error) {
	return []competition.CompetitionJudge{}, nil
}
func (routeJudgeRepository) ListAdminCompetitionJudges(context.Context, int64) ([]competition.CompetitionJudge, error) {
	return []competition.CompetitionJudge{}, nil
}
func (routeJudgeRepository) ValidateCompetitionJudgeAssignment(context.Context, int64, int64, int64) (competition.JudgeAssignmentValidation, error) {
	return competition.JudgeAssignmentValidation{}, nil
}
func (routeJudgeRepository) AssignCompetitionJudge(context.Context, int64, int64, int64) (competition.CreatedCompetitionJudge, error) {
	return competition.CreatedCompetitionJudge{}, nil
}
func (routeJudgeRepository) RemoveCompetitionJudge(context.Context, int64, int64) (bool, error) {
	return true, nil
}
func (routeJudgeRepository) ListCompetitionsByJudgeUser(
	context.Context, int64, int32, int32,
) ([]competition.JudgeCompetition, error) {
	return []competition.JudgeCompetition{}, nil
}
func (routeJudgeRepository) ListJudgeAssignments(
	context.Context, int64, int32, int32,
) ([]competition.JudgeAssignment, error) {
	return []competition.JudgeAssignment{}, nil
}

func TestCompetitionJudgeRoutesAreRegistered(t *testing.T) {
	permissions := map[string]string{}
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions[r.Method] = permission
			handler(w, r)
		})
	}
	mux := http.NewServeMux()
	registerCompetitionJudgeRoutes(
		mux, competition.NewJudgeHandler(competition.NewJudgeService(routeJudgeRepository{})), protected,
	)

	getResponse := httptest.NewRecorder()
	mux.ServeHTTP(getResponse, httptest.NewRequest(http.MethodGet, "/api/v1/competitions/3/judges", nil))
	if getResponse.Code != http.StatusOK {
		t.Fatalf("public GET status = %d", getResponse.Code)
	}
	if _, protected := permissions[http.MethodGet]; protected {
		t.Fatal("GET judges must be public")
	}

	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/api/v1/competitions/3/judges", `{"judge_user_id":42,"position_id":1}`},
		{http.MethodDelete, "/api/v1/competitions/3/judges/8", ""},
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s route status = %d", tc.method, response.Code)
		}
		if permissions[tc.method] != competition.PermissionManageJudges {
			t.Fatalf("%s permission = %q", tc.method, permissions[tc.method])
		}
	}
}

func TestCompetitionJudgeWriteRoutesRequireTokenAndPermission(t *testing.T) {
	manager := token.New("competition-judge-route-secret", time.Minute, time.Hour)
	middleware := rbac.NewMiddleware(manager, denyingRBACRepository{})
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return middleware.Authenticate(middleware.Require(permission)(handler))
	}
	mux := http.NewServeMux()
	registerCompetitionJudgeRoutes(
		mux, competition.NewJudgeHandler(competition.NewJudgeService(routeJudgeRepository{})), protected,
	)
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 42})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/api/v1/competitions/3/judges", `{"judge_user_id":42,"position_id":1}`},
		{http.MethodDelete, "/api/v1/competitions/3/judges/8", ""},
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s without token status = %d", tc.method, response.Code)
		}

		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Authorization", "Bearer "+accessToken)
		response = httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s without permission status = %d", tc.method, response.Code)
		}
	}
}
