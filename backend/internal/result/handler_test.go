package result

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resultRequest(method, body string) *http.Request {
	request := httptest.NewRequest(method, "/", strings.NewReader(body))
	request.SetPathValue("competitionID", "1")
	request.SetPathValue("eventID", "2")
	return request
}

func TestHandlerSerializesNonNilEmptyEntries(t *testing.T) {
	repo := &repositoryStub{contextData: algorithmicContext(), table: Table{
		Sheet:   Sheet{ID: 1, EventID: 2, Mode: ModeIndividual, TaskCount: 3, Status: SheetDraft},
		Entries: []Entry{},
	}}
	response := httptest.NewRecorder()
	NewHandler(NewService(repo, &storageStub{}, &protocolStub{})).Get(
		response, resultRequest(http.MethodGet, ""),
	)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"entries":[]`) ||
		!strings.Contains(response.Body.String(), `"discipline_name":"программирование алгоритмическое"`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandlerRejectsUnknownJSONFields(t *testing.T) {
	repo := &repositoryStub{contextData: algorithmicContext(), table: draftTable()}
	response := httptest.NewRecorder()
	NewHandler(NewService(repo, &storageStub{}, &protocolStub{})).Configure(
		response,
		resultRequest(http.MethodPut, `{"mode":"individual","task_count":2,"unexpected":true}`),
	)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestHandlerHidesRepositoryError(t *testing.T) {
	repo := &repositoryStub{err: errors.New("sensitive database error")}
	response := httptest.NewRecorder()
	NewHandler(NewService(repo, &storageStub{}, &protocolStub{})).Get(
		response, resultRequest(http.MethodGet, ""),
	)
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "sensitive") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}
