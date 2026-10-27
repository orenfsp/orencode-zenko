package main

import (
	"net/http"

	rankuser "github.com/orenfsp/orencode-zenko.git/internal/rank_user"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

func registerRankApplicationRoutes(
	mux *http.ServeMux,
	handler *rankuser.Handler,
	authenticated func(http.HandlerFunc) http.Handler,
	protected protectedHandler,
) {
	mux.Handle("POST /api/v1/rank-applications", authenticated(handler.Create))
	mux.Handle("GET /api/v1/rank-applications/mine", authenticated(handler.ListMine))
	mux.Handle("GET /api/v1/ranks/history", authenticated(handler.ListHistory))
	mux.Handle("GET /api/v1/admin/rank-applications", protected(rbac.PermissionRankApplicationsReview, handler.ListPending))
	mux.Handle("PATCH /api/v1/admin/rank-applications/{applicationID}", protected(rbac.PermissionRankApplicationsReview, handler.Review))
}
