package competition

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

func competitionMux(repo Repository) *http.ServeMux {
	handler := NewHandler(NewService(repo))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/competitions", handler.List)
	mux.HandleFunc("GET /api/v1/competitions/{id}", handler.Get)
	return mux
}

func TestHandlerListUsesDefaultsAndReturnsEmptyList(t *testing.T) {
	repo := &repositoryStub{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions", nil)
	response := httptest.NewRecorder()
	competitionMux(repo).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if repo.seenLimit != DefaultLimit || repo.seenOffset != 0 {
		t.Fatalf("pagination = (%d, %d)", repo.seenLimit, repo.seenOffset)
	}
	var body struct {
		Data []ResponseDTO `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.Data == nil || len(body.Data) != 0 {
		t.Fatalf("data = %#v, want []", body.Data)
	}
}

func TestHandlerRejectsNonNumericPagination(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions?limit=many", nil)
	response := httptest.NewRecorder()
	competitionMux(&repositoryStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestHandlerRejectsOutOfRangePagination(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions?limit=101", nil)
	response := httptest.NewRecorder()
	competitionMux(&repositoryStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestHandlerGetsCompetitionByID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions/7", nil)
	response := httptest.NewRecorder()
	competitionMux(&repositoryStub{item: testCompetition()}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	var body struct {
		Data ResponseDTO `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.Data.ID != 7 {
		t.Fatalf("data.id = %d, want 7", body.Data.ID)
	}
}

func TestHandlerReturnsNotFound(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions/99", nil)
	response := httptest.NewRecorder()
	competitionMux(&repositoryStub{getErr: competitionerrs.ErrNotFound}).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestHandlerRejectsInvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions/not-a-number", nil)
	response := httptest.NewRecorder()
	competitionMux(&repositoryStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestHandlerReturnsInternalErrorForRepositoryFailure(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions", nil)
	response := httptest.NewRecorder()
	competitionMux(&repositoryStub{listErr: errors.New("database unavailable")}).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
}
