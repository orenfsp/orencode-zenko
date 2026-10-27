package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/dashboard"
)

type dashboardServiceStub struct{}

func (dashboardServiceStub) Get(context.Context) (dashboard.Data, error) {
	return dashboard.Data{}, nil
}

func TestRegisterDashboardRoutesUsesPermission(t *testing.T) {
	mux := http.NewServeMux()
	permission := ""
	registerDashboardRoutes(mux, dashboard.NewHandler(dashboardServiceStub{}), func(value string, handler http.HandlerFunc) http.Handler {
		permission = value
		return handler
	})
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil))
	if response.Code != http.StatusOK || permission != dashboard.PermissionRead {
		t.Fatalf("status=%d permission=%q", response.Code, permission)
	}
}
