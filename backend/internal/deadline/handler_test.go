package deadline

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListRequiresIdentity(t *testing.T) {
	handler := NewHandler(NewService(&repositoryStub{}, permissionStub{}))
	response := httptest.NewRecorder()
	handler.List(response, httptest.NewRequest(http.MethodGet, "/api/v1/deadlines", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestCreateReportRejectsUnknownField(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/report-deadlines", strings.NewReader(`{"name":"x","due_at":"2027-02-01","extra":true}`))
	var body CreateReportRequest
	if err := decodeJSON(response, request, &body); err == nil {
		t.Fatal("expected unknown field error")
	}
}
