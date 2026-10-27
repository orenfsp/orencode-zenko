package judgecategory

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRequiresAuthenticatedIdentity(t *testing.T) {
	handler := NewHandler(NewService(&repositoryStub{}, roleStub{}))
	response := httptest.NewRecorder()
	handler.Create(response, httptest.NewRequest(http.MethodPost, "/api/v1/judge-category-applications", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestListPendingRejectsInvalidPagination(t *testing.T) {
	handler := NewHandler(NewService(&repositoryStub{}, roleStub{}))
	response := httptest.NewRecorder()
	handler.ListPending(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/judge-category-applications?limit=0", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", response.Code)
	}
}
