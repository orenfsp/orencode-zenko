package competition

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func judgeRequest(t *testing.T, method, target, competitionID, assignmentID string, body string) *http.Request {
	t.Helper()
	request := authenticatedRequest(t, method, target, bytes.NewReader([]byte(body)))
	request.SetPathValue("id", competitionID)
	if assignmentID != "" {
		request.SetPathValue("assignmentId", assignmentID)
	}
	return request
}

func TestJudgeHandlerListsAndSerializesJudges(t *testing.T) {
	repository := &judgeRepositoryStub{
		status: "published",
		items: []CompetitionJudge{{
			ID: 8, CompetitionID: 3, JudgeUserID: 42, FirstName: "Иван", LastName: "Иванов",
			PositionID: 1, PositionName: "Главный судья",
		}},
	}
	handler := NewJudgeHandler(NewJudgeService(repository))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitions/3/judges", nil)
	request.SetPathValue("id", "3")
	response := httptest.NewRecorder()
	handler.List(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"judge":{"id":42,"first_name":"Иван"`) ||
		!strings.Contains(response.Body.String(), `"position":{"id":1,"name":"Главный судья"}`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestJudgeHandlerAssignsAndRemovesJudge(t *testing.T) {
	repository := &judgeRepositoryStub{
		status: "draft", validation: validJudgeValidation(), removed: true,
		created: CreatedCompetitionJudge{ID: 8, CompetitionID: 3, JudgeUserID: 42, PositionID: 1},
	}
	handler := NewJudgeHandler(NewJudgeService(repository))

	assignResponse := httptest.NewRecorder()
	handler.Assign(assignResponse, judgeRequest(
		t, http.MethodPost, "/api/v1/competitions/3/judges", "3", "", `{"judge_user_id":42,"position_id":1}`,
	))
	if assignResponse.Code != http.StatusCreated || !strings.Contains(assignResponse.Body.String(), `"judge_user_id":42`) {
		t.Fatalf("assign response = %d %s", assignResponse.Code, assignResponse.Body.String())
	}

	removeResponse := httptest.NewRecorder()
	handler.Remove(removeResponse, judgeRequest(
		t, http.MethodDelete, "/api/v1/competitions/3/judges/8", "3", "8", "",
	))
	if removeResponse.Code != http.StatusNoContent || removeResponse.Body.Len() != 0 {
		t.Fatalf("remove response = %d %s", removeResponse.Code, removeResponse.Body.String())
	}
}

func TestJudgeHandlerRejectsMalformedJSONAndUnknownFields(t *testing.T) {
	handler := NewJudgeHandler(NewJudgeService(&judgeRepositoryStub{status: "draft"}))
	for _, body := range []string{`{"judge_user_id":`, `{"judge_user_id":42,"position_id":1,"id":7}`} {
		response := httptest.NewRecorder()
		handler.Assign(response, judgeRequest(t, http.MethodPost, "/", "3", "", body))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d", body, response.Code)
		}
	}
}

func TestJudgeHandlerHidesPrivateCompetition(t *testing.T) {
	handler := NewJudgeHandler(NewJudgeService(&judgeRepositoryStub{status: "draft"}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.SetPathValue("id", "3")
	response := httptest.NewRecorder()
	handler.List(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d", response.Code)
	}
}
