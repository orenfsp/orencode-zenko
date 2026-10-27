package competition

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type handlerRBACRepository struct{}

func (handlerRBACRepository) HasPermission(context.Context, int64, string) (bool, error) {
	return true, nil
}
func (handlerRBACRepository) CreateRole(context.Context, rbac.CreateRoleRequest) (rbac.Role, error) {
	return rbac.Role{}, nil
}
func (handlerRBACRepository) AssignRole(context.Context, int64, int64) error { return nil }
func (handlerRBACRepository) ListRoles(context.Context) ([]rbac.Role, error) { return nil, nil }
func (handlerRBACRepository) ListPermissions(context.Context) ([]rbac.Permission, error) {
	return nil, nil
}
func (handlerRBACRepository) ListSelfAssignableRoles(context.Context) ([]rbac.Role, error) {
	return nil, nil
}

func authenticatedRequest(t *testing.T, method, target string, body *bytes.Reader) *http.Request {
	t.Helper()
	manager := token.New("competition-handler-test-secret", time.Minute, time.Hour)
	accessToken, err := manager.GenerateAccessToken(token.TokenPayload{UserID: 42})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	var captured *http.Request
	rbac.NewMiddleware(manager, handlerRBACRepository{}).Authenticate(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r
	})).ServeHTTP(httptest.NewRecorder(), request)
	if captured == nil {
		t.Fatal("authentication middleware did not pass the request")
	}
	return captured
}

func competitionMux(repo Repository) *http.ServeMux {
	handler := NewHandler(NewService(repo))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/calendar/competitions", handler.Calendar)
	mux.HandleFunc("GET /api/v1/competitions", handler.List)
	mux.HandleFunc("GET /api/v1/competitions/{id}", handler.Get)
	return mux
}

func TestHandlerCalendarIsPublicAndUsesDateOnlyBoundaries(t *testing.T) {
	repo := &repositoryStub{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/calendar/competitions?from=2026-11-01&to=2026-12-01", nil)
	response := httptest.NewRecorder()
	competitionMux(repo).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	wantFrom := time.Date(2026, time.November, 1, 0, 0, 0, 0, calendarLocation)
	wantTo := time.Date(2026, time.December, 1, 0, 0, 0, 0, calendarLocation)
	if !repo.seenCalendar.From.Equal(wantFrom) || !repo.seenCalendar.To.Equal(wantTo) {
		t.Fatalf("calendar period = %s..%s, want %s..%s", repo.seenCalendar.From, repo.seenCalendar.To, wantFrom, wantTo)
	}
	if repo.seenCalendar.From.Location().String() != "Asia/Yekaterinburg" ||
		repo.seenCalendar.To.Location().String() != "Asia/Yekaterinburg" {
		t.Fatalf("date-only boundaries must use Asia/Yekaterinburg: %s..%s", repo.seenCalendar.From, repo.seenCalendar.To)
	}
	if repo.seenCalendar.Limit != DefaultLimit || repo.seenCalendar.Offset != 0 {
		t.Fatalf("pagination = (%d, %d)", repo.seenCalendar.Limit, repo.seenCalendar.Offset)
	}
	var body struct {
		Data []CalendarCompetitionDTO `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.Data == nil || len(body.Data) != 0 {
		t.Fatalf("data = %#v, want []", body.Data)
	}
}

func TestHandlerCalendarPreservesRFC3339OffsetsAndParsesFilters(t *testing.T) {
	repo := &repositoryStub{}
	target := "/api/v1/calendar/competitions?from=2026-02-01T00%3A00%3A00%2B05%3A00&to=2026-03-01T00%3A00%3A00%2B05%3A00&discipline_id=2&level_id=4&limit=25&offset=10"
	response := httptest.NewRecorder()
	competitionMux(repo).ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	wantFrom := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.FixedZone("", 5*60*60))
	if !repo.seenCalendar.From.Equal(wantFrom) || repo.seenCalendar.From.Format("-07:00") != "+05:00" {
		t.Fatalf("from = %s, want instant %s with +05:00 offset", repo.seenCalendar.From, wantFrom)
	}
	if repo.seenCalendar.DisciplineID == nil || *repo.seenCalendar.DisciplineID != 2 ||
		repo.seenCalendar.LevelID == nil || *repo.seenCalendar.LevelID != 4 ||
		repo.seenCalendar.Limit != 25 || repo.seenCalendar.Offset != 10 {
		t.Fatalf("calendar filter = %#v", repo.seenCalendar)
	}
}

func TestHandlerCalendarReturnsCompactJSON(t *testing.T) {
	item := testCompetition()
	item.Description = stringPointer("not for calendar")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/calendar/competitions?from=2026-11-01&to=2026-12-01", nil)
	competitionMux(&repositoryStub{calendarItems: []Competition{item}}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var body map[string][]map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(body["data"]) != 1 {
		t.Fatalf("data = %#v", body["data"])
	}
	for _, field := range []string{"description", "registration_starts_at", "registration_ends_at", "calendar_plan", "regulation_url"} {
		if _, exists := body["data"][0][field]; exists {
			t.Fatalf("compact response unexpectedly contains %q", field)
		}
	}
}

func TestHandlerCalendarRejectsInvalidQuery(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		status int
	}{
		{name: "missing period", query: "?from=2026-01-01", status: http.StatusBadRequest},
		{name: "date without timezone", query: "?from=2026-01-01T12:00:00&to=2026-01-02T12:00:00Z", status: http.StatusBadRequest},
		{name: "empty period", query: "?from=2026-02-01&to=2026-02-01", status: http.StatusBadRequest},
		{name: "reversed period", query: "?from=2026-02-02&to=2026-02-01", status: http.StatusBadRequest},
		{name: "invalid discipline", query: "?from=2026-01-01&to=2026-02-01&discipline_id=0", status: http.StatusBadRequest},
		{name: "invalid level", query: "?from=2026-01-01&to=2026-02-01&level_id=many", status: http.StatusBadRequest},
		{name: "invalid limit", query: "?from=2026-01-01&to=2026-02-01&limit=101", status: http.StatusBadRequest},
		{name: "invalid offset", query: "?from=2026-01-01&to=2026-02-01&offset=-1", status: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/calendar/competitions"+tt.query, nil)
			competitionMux(&repositoryStub{}).ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.status, response.Body.String())
			}
		})
	}
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

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
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
	request := authenticatedRequest(t, http.MethodPost, "/api/v1/competitions", bytes.NewReader(body))
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
	request := authenticatedRequest(t, http.MethodPost, "/api/v1/competitions", bytes.NewReader([]byte(`{"name":`)))
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
	request := authenticatedRequest(t, http.MethodPost, "/api/v1/competitions", bytes.NewReader(body))
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{})).Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestHandlerCreateReturnsConflictForDuplicateName(t *testing.T) {
	body, _ := json.Marshal(validCreateRequest())
	request := authenticatedRequest(t, http.MethodPost, "/api/v1/competitions", bytes.NewReader(body))
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{createErr: competitionerrs.ErrNameConflict})).Create(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", response.Code)
	}
}

func TestHandlerUpdatesCompetitionPartially(t *testing.T) {
	current := testCompetition()
	current.Status = "draft"
	current.Description = stringPointer("Сохранить")
	updated := current
	updated.Name = "Новое название"
	repo := &repositoryStub{item: current, created: updated}
	request := authenticatedRequest(t, http.MethodPatch, "/api/v1/competitions/7", bytes.NewReader([]byte(`{"name":"Новое название"}`)))
	request.SetPathValue("id", "7")
	response := httptest.NewRecorder()

	NewHandler(NewService(repo)).Update(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if repo.seenUpdate.Name != "Новое название" || repo.seenUpdate.Description == nil || *repo.seenUpdate.Description != "Сохранить" {
		t.Fatalf("repository update = %#v", repo.seenUpdate)
	}
}

func TestHandlerUpdateDistinguishesNullFromOmitted(t *testing.T) {
	current := testCompetition()
	current.Status = "draft"
	current.Description = stringPointer("Удалить")
	updated := current
	updated.Description = nil
	repo := &repositoryStub{item: current, created: updated}
	request := authenticatedRequest(t, http.MethodPatch, "/api/v1/competitions/7", bytes.NewReader([]byte(`{"description":null}`)))
	request.SetPathValue("id", "7")
	response := httptest.NewRecorder()

	NewHandler(NewService(repo)).Update(response, request)

	if response.Code != http.StatusOK || repo.seenUpdate.Description != nil {
		t.Fatalf("status = %d, description = %#v", response.Code, repo.seenUpdate.Description)
	}
}

func TestHandlerUpdateRejectsInvalidPayloads(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty patch", body: `{}`},
		{name: "unknown field", body: `{"status":"published"}`},
		{name: "null required field", body: `{"name":null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := testCompetition()
			current.Status = "draft"
			request := authenticatedRequest(t, http.MethodPatch, "/api/v1/competitions/7", bytes.NewReader([]byte(tt.body)))
			request.SetPathValue("id", "7")
			response := httptest.NewRecorder()

			NewHandler(NewService(&repositoryStub{item: current})).Update(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestHandlerUpdateMapsDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		repo   *repositoryStub
		status int
	}{
		{name: "missing", repo: &repositoryStub{getErr: competitionerrs.ErrNotFound}, status: http.StatusNotFound},
		{name: "not draft", repo: &repositoryStub{item: testCompetition()}, status: http.StatusConflict},
		{name: "repository failure", repo: &repositoryStub{getErr: errors.New("database unavailable")}, status: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := authenticatedRequest(t, http.MethodPatch, "/api/v1/competitions/7", bytes.NewReader([]byte(`{"name":"Новое"}`)))
			request.SetPathValue("id", "7")
			response := httptest.NewRecorder()

			NewHandler(NewService(tt.repo)).Update(response, request)

			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
		})
	}
}
