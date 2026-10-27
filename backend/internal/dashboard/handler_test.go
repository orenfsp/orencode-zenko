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
}

func (s serviceStub) Get(context.Context) (Data, error) { return s.data, s.err }

func TestHandlerGet(t *testing.T) {
	handler := NewHandler(serviceStub{data: normalize(Data{Metrics: Metrics{Competitions: 4}})})
	response := httptest.NewRecorder()
	handler.Get(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"competitions":4`) || !strings.Contains(response.Body.String(), `"deadlines":[]`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestHandlerGetRepositoryError(t *testing.T) {
	handler := NewHandler(serviceStub{err: errors.New("boom")})
	response := httptest.NewRecorder()
	handler.Get(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "boom") {
		t.Fatalf("internal error leaked: %s", response.Body.String())
	}
}
