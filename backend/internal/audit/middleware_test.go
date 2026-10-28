package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

func TestMiddlewareWritesSuccessfulRequest(t *testing.T) {
	repo := &repositoryStub{}
	tokens := token.New("secret", time.Hour, time.Hour)
	accessToken, err := tokens.GenerateAccessToken(token.TokenPayload{UserID: 12})
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/48/roles", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()
	handler := NewMiddleware(repo, tokens).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(response, request)

	if len(repo.written) != 1 {
		t.Fatalf("written events = %d", len(repo.written))
	}
	event := repo.written[0]
	if event.ActorUserID == nil || *event.ActorUserID != 12 || event.EntityID == nil || *event.EntityID != 48 || event.Status != 204 {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestMiddlewareWritesFailedRequest(t *testing.T) {
	repo := &repositoryStub{}
	tokens := token.New("secret", time.Hour, time.Hour)
	accessToken, err := tokens.GenerateAccessToken(token.TokenPayload{UserID: 12})
	if err != nil {
		t.Fatal(err)
	}
	handler := NewMiddleware(repo, tokens).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/profile", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)

	handler.ServeHTTP(httptest.NewRecorder(), request)

	if len(repo.written) != 1 {
		t.Fatalf("written events = %d", len(repo.written))
	}
	event := repo.written[0]
	if event.ActorUserID == nil || *event.ActorUserID != 12 || event.Status != http.StatusBadRequest {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestMiddlewareWritesFailedLoginWithoutActor(t *testing.T) {
	repo := &repositoryStub{}
	tokens := token.New("secret", time.Hour, time.Hour)
	handler := NewMiddleware(repo, tokens).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil))

	if len(repo.written) != 1 || repo.written[0].ActorUserID != nil || repo.written[0].EventType != EventLogin {
		t.Fatalf("unexpected events: %+v", repo.written)
	}
}

func TestMiddlewareGetsActorFromSuccessfulLoginResponse(t *testing.T) {
	repo := &repositoryStub{}
	tokens := token.New("secret", time.Hour, time.Hour)
	accessToken, err := tokens.GenerateAccessToken(token.TokenPayload{UserID: 27})
	if err != nil {
		t.Fatal(err)
	}
	handler := NewMiddleware(repo, tokens).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]string{"access_token": accessToken}})
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil))

	if len(repo.written) != 1 || repo.written[0].ActorUserID == nil || *repo.written[0].ActorUserID != 27 {
		t.Fatalf("unexpected events: %+v", repo.written)
	}
}

func TestMiddlewareGetsActorFromLogoutRefreshTokenAndRestoresBody(t *testing.T) {
	repo := &repositoryStub{}
	tokens := token.New("secret", time.Hour, time.Hour)
	refreshToken, err := tokens.GenerateRefreshToken(token.TokenPayload{UserID: 31})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	handler := NewMiddleware(repo, tokens).Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]string
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request["refresh_token"] == "" {
			t.Fatalf("request body was not restored: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(body)))

	if len(repo.written) != 1 || repo.written[0].ActorUserID == nil || *repo.written[0].ActorUserID != 31 {
		t.Fatalf("unexpected events: %+v", repo.written)
	}
}

func TestAuditedRoutes(t *testing.T) {
	tests := []struct {
		method, path, eventType string
	}{
		{http.MethodPost, "/api/v1/competitions", EventCompetitionCreated},
		{http.MethodPatch, "/api/v1/competitions/10", EventCompetitionUpdated},
		{http.MethodPost, "/api/v1/admin/competitions/10/publish", EventCompetitionPublished},
		{http.MethodPatch, "/api/v1/competitions/10/events/20", EventCompetitionEventUpdated},
		{http.MethodDelete, "/api/v1/competitions/10/judges/30", EventCompetitionJudgeRemoved},
		{http.MethodPost, "/api/v1/admin/competition-applications/40/reject", EventCompetitionApplicationReject},
		{http.MethodPatch, "/api/v1/admin/rank-applications/50", EventRankApplicationReviewed},
		{http.MethodPost, "/api/v1/admin/competitions/10/events/20/results/approve", EventResultApproved},
		{http.MethodPut, "/api/v1/admin/competitions/10/events/20/protocol", EventProtocolUploaded},
		{http.MethodPut, "/api/v1/admin/rating/selections", EventRatingSelectionSaved},
		{http.MethodPost, "/api/v1/admin/rating/selections/60/lock", EventRatingSelectionLocked},
	}
	for _, tt := range tests {
		rule, ok := auditedRoute(tt.method, tt.path)
		if !ok || rule.eventType != tt.eventType {
			t.Errorf("auditedRoute(%q, %q) = %+v, %v", tt.method, tt.path, rule, ok)
		}
	}
}

func TestMiddlewareSkipsUnauditedRoute(t *testing.T) {
	repo := &repositoryStub{}
	tokens := token.New("secret", time.Hour, time.Hour)
	handler := NewMiddleware(repo, tokens).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/competitions", nil))

	if len(repo.written) != 0 {
		t.Fatalf("written events = %d", len(repo.written))
	}
}
