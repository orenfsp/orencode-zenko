package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/antidoping"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

func registerAntiDopingRoutes(
	mux *http.ServeMux,
	handler *antidoping.Handler,
	authenticated func(http.HandlerFunc) http.Handler,
	protected func(string, http.HandlerFunc) http.Handler,
) {
	mux.Handle("POST /api/v1/anti-doping/self", authenticated(handler.Upload))
	mux.Handle("GET /api/v1/anti-doping/self", authenticated(handler.ListMine))
	mux.Handle(
		"PATCH /api/v1/admin/anti-doping/{certificateID}",
		protected(rbac.PermissionAntiDopingReview, handler.Review),
	)
}
