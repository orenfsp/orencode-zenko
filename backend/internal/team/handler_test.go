package team

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	usercoach "github.com/orenfsp/orencode-zenko.git/internal/user_coach"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type handlerRepositoryStub struct{}

func (handlerRepositoryStub) Create(context.Context, string, []int64, []int64) (Squad, error) {
	return Squad{}, nil
}
func (handlerRepositoryStub) List(context.Context) ([]Squad, error) { return nil, nil }
func (handlerRepositoryStub) ListByCoach(context.Context, int64) ([]Squad, error) {
	return nil, nil
}

type handlerRoleCheckerStub struct{}

func (handlerRoleCheckerStub) HasRole(_ context.Context, _ int64, role string) (bool, error) {
	return role == "coach", nil
}

type pendingRelationsStub struct{}

func (pendingRelationsStub) ListByCoach(context.Context, int64, int32, int32) ([]usercoach.UserCoach, error) {
	return []usercoach.UserCoach{{UserID: 20, Status: usercoach.StatusPending}}, nil
}

func TestCreateReturnsForbiddenForUnconfirmedAthlete(t *testing.T) {
	service := NewService(handlerRepositoryStub{}, handlerRoleCheckerStub{}, pendingRelationsStub{})
	handler := NewHandler(service)
	tokens := token.New("test-secret", time.Minute, time.Hour)
	accessToken, err := tokens.GenerateAccessToken(token.TokenPayload{UserID: 10})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/teams", strings.NewReader(
		`{"name":"Сборная","member_ids":[20],"coach_ids":[10]}`,
	))
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()
	rbac.NewMiddleware(tokens, nil).Authenticate(http.HandlerFunc(handler.Create)).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusForbidden, response.Body.String())
	}
}
