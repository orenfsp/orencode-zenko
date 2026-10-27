package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/deadline"
)

func registerDeadlineRoutes(
	mux *http.ServeMux,
	handler *deadline.Handler,
	authenticated func(http.HandlerFunc) http.Handler,
	protected protectedHandler,
) {
	mux.Handle("GET /api/v1/deadlines", authenticated(handler.List))
	mux.Handle("GET /api/v1/admin/report-deadlines", protected(deadline.PermissionManage, handler.ListReports))
	mux.Handle("POST /api/v1/admin/report-deadlines", protected(deadline.PermissionManage, handler.CreateReport))
	mux.Handle("PATCH /api/v1/admin/report-deadlines/{id}", protected(deadline.PermissionManage, handler.UpdateReport))
}
