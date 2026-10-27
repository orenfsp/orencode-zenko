package registry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerListUsersSerializesPage(t *testing.T) {
	handler := NewHandler(NewService(&repositoryStub{users: Page[User]{Items: []User{}, Total: 0, Limit: 50}}))
	response := httptest.NewRecorder()
	handler.ListUsers(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/registry/users", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[]`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestHandlerListUsersRejectsInvalidFilter(t *testing.T) {
	handler := NewHandler(NewService(&repositoryStub{}))
	response := httptest.NewRecorder()
	handler.ListUsers(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/registry/users?rank_id=no", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", response.Code)
	}
}
