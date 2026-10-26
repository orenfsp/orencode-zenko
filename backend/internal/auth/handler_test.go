package auth

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRefreshRejectsMissingRefreshToken(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()
	(&Handler{}).Refresh(rr, r)

	if rr.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%q", rr.Code, rr.Body.String())
	}
	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if response.Message == "" {
		t.Fatal("validation error message is empty")
	}
}

func TestLoginRejectsWrongJSONType(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"user_name":123,"password":"p","device_id":"d"}`))
	rr := httptest.NewRecorder()
	(&Handler{}).Login(rr, r)

	if rr.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%q", rr.Code, rr.Body.String())
	}
}
