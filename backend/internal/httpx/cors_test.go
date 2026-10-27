package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testAllowedOrigin = "http://localhost:5173"

// newCORSHandler оборачивает next в CORS с заданным allowlist.
// Если next == nil, подставляется «сторожевой» обработчик, который валит
// тест при вызове — это полезно для preflight-тестов, где next вызываться
// не должен.
func newCORSHandler(t *testing.T, allowed []string, next http.Handler) http.Handler {
	t.Helper()
	if next == nil {
		next = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("next handler must not be called")
		})
	}
	return CORS(allowed)(next)
}

func TestCORSPreflightWithAllowedOrigin(t *testing.T) {
	var nextCalled bool
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	handler := newCORSHandler(t, []string{testAllowedOrigin}, next)

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/profile", nil)
	request.Header.Set("Origin", testAllowedOrigin)
	request.Header.Set("Access-Control-Request-Method", "PATCH")
	request.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if nextCalled {
		t.Fatal("preflight request must not reach the next handler")
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != testAllowedOrigin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, testAllowedOrigin)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("Access-Control-Allow-Methods is empty")
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("Access-Control-Allow-Headers is empty")
	}
	if got := response.Header().Get("Access-Control-Max-Age"); got == "" {
		t.Fatal("Access-Control-Max-Age is empty")
	}
	if got := response.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary = %q, want Origin", got)
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	// next == nil → сторожевой обработчик; preflight не должен его звать.
	handler := newCORSHandler(t, []string{testAllowedOrigin}, nil)

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/profile", nil)
	request.Header.Set("Origin", "https://evil.example.com")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got != "" {
		t.Fatalf("Access-Control-Allow-Methods = %q, want empty", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got != "" {
		t.Fatalf("Access-Control-Allow-Headers = %q, want empty", got)
	}
}

func TestCORSPassesThroughNonPreflightRequests(t *testing.T) {
	var nextCalled bool
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	})
	handler := newCORSHandler(t, []string{testAllowedOrigin}, next)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	request.Header.Set("Origin", testAllowedOrigin)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if !nextCalled {
		t.Fatal("non-preflight request must reach the next handler")
	}
	if response.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != testAllowedOrigin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, testAllowedOrigin)
	}
}

func TestCORSPreflightWithoutOriginHeader(t *testing.T) {
	handler := newCORSHandler(t, []string{testAllowedOrigin}, nil)

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/profile", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCORSWithEmptyAllowlist(t *testing.T) {
	// Пустой allowlist не должен ничего разрешать.
	handler := newCORSHandler(t, nil, nil)

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/profile", nil)
	request.Header.Set("Origin", testAllowedOrigin)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}
