package reference

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerSerializesReferenceDTOs(t *testing.T) {
	code, levelName := "regional", "Региональный"
	minAge, maxAge := int16(14), int16(17)
	handler := NewHandler(NewService(&repositoryStub{
		levels:    []CompetitionLevel{{ID: 3, Code: &code, Name: &levelName}},
		ageGroups: []AgeGroup{{ID: 4, Name: "Юниоры", MinAge: &minAge, MaxAge: &maxAge}},
	}))

	levelResponse := httptest.NewRecorder()
	handler.ListCompetitionLevels(levelResponse, httptest.NewRequest(http.MethodGet, "/", nil))
	if levelResponse.Code != http.StatusOK || levelResponse.Header().Get("Content-Type") != "application/json" ||
		!strings.Contains(levelResponse.Body.String(), `"data":[{"id":3,"code":"regional","name":"Региональный"}]`) {
		t.Fatalf("levels response = %d %s", levelResponse.Code, levelResponse.Body.String())
	}

	ageResponse := httptest.NewRecorder()
	handler.ListAgeGroups(ageResponse, httptest.NewRequest(http.MethodGet, "/", nil))
	if ageResponse.Code != http.StatusOK ||
		!strings.Contains(ageResponse.Body.String(), `"data":[{"id":4,"name":"Юниоры","min_age":14,"max_age":17}]`) {
		t.Fatalf("age groups response = %d %s", ageResponse.Code, ageResponse.Body.String())
	}
}

func TestHandlerSerializesEmptyListAsArray(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{})).ListDisciplines(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"data":[]`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandlerHidesRepositoryError(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler(NewService(&repositoryStub{err: errors.New("sensitive database error")})).ListDisciplines(
		response, httptest.NewRequest(http.MethodGet, "/", nil),
	)
	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "sensitive") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}
