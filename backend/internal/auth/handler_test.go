package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRefreshRejectsMissingRefreshToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(`{}`))
	response := httptest.NewRecorder()
	(&Handler{}).Refresh(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestLoginRejectsWrongJSONType(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":123}`))
	response := httptest.NewRecorder()
	(&Handler{}).Login(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	var body ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
}

func TestPatchProfileRequiresBearerToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/profile", strings.NewReader(`{"tg_username":"username"}`))
	response := httptest.NewRecorder()
	(&Handler{}).PatchProfile(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}
