package competition

import (
	"bytes"
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

func TestHandlerCreateRequiresAuthorizedEmployee(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitions", bytes.NewReader([]byte(`{}`)))
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{})).Create(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestHandlerCreatesCompetition(t *testing.T) {
	req := validCreateRequest()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	created := testCompetition()
	created.Status = "draft"
	repo := &repositoryStub{created: created}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitions", bytes.NewReader(body))
	request = request.WithContext(WithAuthorizedEmployee(request.Context(), 42))
	response := httptest.NewRecorder()
	NewHandler(NewService(repo)).Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	if repo.seenCreate.CreatedBy != 42 {
		t.Fatalf("created_by = %d, want 42", repo.seenCreate.CreatedBy)
	}
	var responseBody struct {
		Data ResponseDTO `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if responseBody.Data.Status != "draft" {
		t.Fatalf("status in response = %q, want draft", responseBody.Data.Status)
	}
}

func TestHandlerCreateRejectsInvalidJSON(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitions", bytes.NewBufferString(`{"name":`))
	request = request.WithContext(WithAuthorizedEmployee(request.Context(), 42))
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{})).Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestHandlerCreateRejectsInvalidData(t *testing.T) {
	req := validCreateRequest()
	req.Name = ""
	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitions", bytes.NewReader(body))
	request = request.WithContext(WithAuthorizedEmployee(request.Context(), 42))
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{})).Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestHandlerCreateReturnsConflictForDuplicateName(t *testing.T) {
	body, _ := json.Marshal(validCreateRequest())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitions", bytes.NewReader(body))
	request = request.WithContext(WithAuthorizedEmployee(request.Context(), 42))
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{createErr: competitionerrs.ErrNameConflict})).Create(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", response.Code)
	}
}
