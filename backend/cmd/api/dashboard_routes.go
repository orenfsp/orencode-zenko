package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/dashboard"
)

func registerDashboardRoutes(mux *http.ServeMux, handler *dashboard.Handler, protected protectedHandler) {
	mux.Handle("GET /api/v1/admin/dashboard", protected(dashboard.PermissionRead, handler.Get))
}
