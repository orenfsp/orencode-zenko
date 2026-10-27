package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/competition"
)

type protectedHandler func(permission string, handler http.HandlerFunc) http.Handler

func registerCompetitionRoutes(mux *http.ServeMux, handler *competition.Handler, protected protectedHandler) {
	mux.HandleFunc("GET /api/v1/calendar/competitions", handler.Calendar)
	mux.HandleFunc("GET /api/v1/competitions", handler.List)
	mux.HandleFunc("GET /api/v1/competitions/{id}", handler.Get)
	mux.Handle("POST /api/v1/competitions", protected(competition.PermissionCreate, handler.Create))
	mux.Handle("PATCH /api/v1/competitions/{id}", protected(competition.PermissionUpdate, handler.Update))
}
