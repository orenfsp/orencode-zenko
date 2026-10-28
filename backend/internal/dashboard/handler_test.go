package dashboard

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type serviceStub struct {
	data Data
	err  error
	year int
}

func (s *serviceStub) Get(_ context.Context, year int) (Data, error) {
	s.year = year
	return s.data, s.err
}

func TestHandlerGet(t *testing.T) {
	service := &serviceStub{data: normalize(Data{Metrics: Metrics{Competitions: 4}})}
	handler := NewHandler(service)
	response := httptest.NewRecorder()
	handler.Get(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard?year=2024", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"competitions":4`) || !strings.Contains(response.Body.String(), `"deadlines":[]`) {
		t.Fatalf("body = %s", response.Body.String())
	}
	if service.year != 2024 {
		t.Fatalf("year = %d, want 2024", service.year)
	}
}

func TestHandlerGetRepositoryError(t *testing.T) {
	handler := NewHandler(&serviceStub{err: errors.New("boom")})
	response := httptest.NewRecorder()
	handler.Get(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "boom") {
		t.Fatalf("internal error leaked: %s", response.Body.String())
	}
}

func TestHandlerRejectsInvalidYear(t *testing.T) {
	handler := NewHandler(&serviceStub{})
	response := httptest.NewRecorder()
	handler.Get(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard?year=nope", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}
