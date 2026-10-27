package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/judgecategory"
)

func registerJudgeCategoryRoutes(
	mux *http.ServeMux,
	handler *judgecategory.Handler,
	authenticated func(http.HandlerFunc) http.Handler,
	protected protectedHandler,
) {
	mux.Handle("POST /api/v1/judge-category-applications", authenticated(handler.Create))
	mux.Handle("GET /api/v1/judge-category-applications/mine", authenticated(handler.ListMine))
	mux.Handle("GET /api/v1/judge-category/current", authenticated(handler.GetCurrent))
	mux.Handle("GET /api/v1/judge/competitions", authenticated(handler.ListAssignments))
	mux.Handle("GET /api/v1/admin/judge-category-applications", protected(judgecategory.PermissionReview, handler.ListPending))
	mux.Handle("PATCH /api/v1/admin/judge-category-applications/{id}", protected(judgecategory.PermissionReview, handler.Review))
}
