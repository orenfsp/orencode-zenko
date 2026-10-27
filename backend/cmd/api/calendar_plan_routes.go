package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/calendarplan"
)

func registerCalendarPlanRoutes(mux *http.ServeMux, handler *calendarplan.Handler, protected protectedHandler) {
	mux.Handle("GET /api/v1/admin/calendar-plans", protected(calendarplan.PermissionRead, handler.List))
	mux.Handle("GET /api/v1/admin/calendar-plans/{id}", protected(calendarplan.PermissionRead, handler.Get))
	mux.Handle("POST /api/v1/admin/calendar-plans", protected(calendarplan.PermissionCreate, handler.Create))
	mux.Handle("PATCH /api/v1/admin/calendar-plans/{id}", protected(calendarplan.PermissionUpdate, handler.Update))
	mux.Handle("GET /api/v1/admin/calendar-plans/{id}/competitions", protected(calendarplan.PermissionRead, handler.ListCompetitions))
}
