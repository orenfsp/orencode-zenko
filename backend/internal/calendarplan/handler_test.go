package calendarplan

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestHandler(repo Repository) *Handler {
	return NewHandler(NewService(repo))
}

func requestWithPathValue(method, target, id string, body *bytes.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	request.SetPathValue("id", id)
	return request
}

func TestHandlerListsPlansWithDefaultsAndYear(t *testing.T) {
	year := int16(2026)
	repo := &repositoryStub{plans: []CalendarPlan{{ID: 1, Name: "План", Year: year}}}
	response := httptest.NewRecorder()
	newTestHandler(repo).List(response, httptest.NewRequest(http.MethodGet, "/?year=2026", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"data":[{"id":1`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
	if repo.listFilter.Year == nil || *repo.listFilter.Year != year || repo.listFilter.Limit != DefaultLimit || repo.listFilter.Offset != 0 {
		t.Fatalf("filter = %#v", repo.listFilter)
	}
}

func TestHandlerListParsesPagination(t *testing.T) {
	repo := &repositoryStub{}
	response := httptest.NewRecorder()
	newTestHandler(repo).List(response, httptest.NewRequest(http.MethodGet, "/?limit=10&offset=3", nil))
	if response.Code != http.StatusOK || repo.listFilter.Limit != 10 || repo.listFilter.Offset != 3 {
		t.Fatalf("response = %d, filter = %#v", response.Code, repo.listFilter)
	}
}

func TestHandlerGetsPlanAndReturnsNotFound(t *testing.T) {
	handler := newTestHandler(&repositoryStub{getPlan: CalendarPlan{ID: 5, Name: "ЕКП", Year: 2026}})
	response := httptest.NewRecorder()
	handler.Get(response, requestWithPathValue(http.MethodGet, "/5", "5", bytes.NewReader(nil)))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"id":5`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	newTestHandler(&repositoryStub{err: ErrNotFound}).Get(response, requestWithPathValue(http.MethodGet, "/6", "6", bytes.NewReader(nil)))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestHandlerCreatesPlan(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":" План ","year":2026}`))
	newTestHandler(&repositoryStub{}).Create(response, request)
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"name":"План"`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandlerRejectsInvalidCreatePayloads(t *testing.T) {
	for _, body := range []string{
		`{"name":"","year":2026}`,
		`{"name":"План","year":0}`,
		`{"name":"План","year":40000}`,
		`{"name":"План","year":2026,"id":1}`,
		`{"name":`,
		`{"name":"` + strings.Repeat("я", MaxNameRunes+1) + `","year":2026}`,
	} {
		response := httptest.NewRecorder()
		newTestHandler(&repositoryStub{}).Create(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d", body, response.Code)
		}
	}
}

func TestHandlerUpdatesPartiallyAndRejectsForbiddenFields(t *testing.T) {
	repo := &repositoryStub{getPlan: CalendarPlan{ID: 3, Name: "План", Year: 2025}}
	response := httptest.NewRecorder()
	request := requestWithPathValue(http.MethodPatch, "/3", "3", bytes.NewReader([]byte(`{"year":2026}`)))
	newTestHandler(repo).Update(response, request)
	if response.Code != http.StatusOK || repo.updated.Name != "План" || repo.updated.Year != 2026 {
		t.Fatalf("response = %d %s, updated = %#v", response.Code, response.Body.String(), repo.updated)
	}

	for _, body := range []string{`{}`, `{"id":7}`, `{"name":null}`, `{"year":null}`} {
		response = httptest.NewRecorder()
		request = requestWithPathValue(http.MethodPatch, "/3", "3", bytes.NewReader([]byte(body)))
		newTestHandler(repo).Update(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %s status = %d", body, response.Code)
		}
	}
}

func TestHandlerListsPlanCompetitions(t *testing.T) {
	repo := &repositoryStub{getPlan: CalendarPlan{ID: 2}, competitions: []Competition{{ID: 8, Name: "Черновик", Status: "draft", DisciplineID: 1, DisciplineName: "ИИ"}}}
	response := httptest.NewRecorder()
	request := requestWithPathValue(http.MethodGet, "/2/competitions?limit=5&offset=2", "2", bytes.NewReader(nil))
	newTestHandler(repo).ListCompetitions(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"draft"`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandlerMapsRepositoryFailureToInternalError(t *testing.T) {
	repo := &repositoryStub{err: errors.New("sensitive database error")}
	response := httptest.NewRecorder()
	newTestHandler(repo).List(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "sensitive") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandlerAlwaysUsesDataArrayForEmptyLists(t *testing.T) {
	response := httptest.NewRecorder()
	newTestHandler(&repositoryStub{}).List(response, httptest.NewRequest(http.MethodGet, "/", nil))
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if string(payload["data"]) != "[]" {
		t.Fatalf("data = %s", payload["data"])
	}
}
