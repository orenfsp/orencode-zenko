package audit

import (
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
	if event.ActorUserID != 12 || event.EntityID == nil || *event.EntityID != 48 || event.Status != 204 {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestMiddlewareSkipsFailedRequest(t *testing.T) {
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

	if len(repo.written) != 0 {
		t.Fatalf("written events = %d", len(repo.written))
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
