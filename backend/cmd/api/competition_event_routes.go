package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/competitionevent"
)

func registerCompetitionEventRoutes(mux *http.ServeMux, handler *competitionevent.Handler, protected protectedHandler) {
	mux.HandleFunc("GET /api/v1/competitions/{competitionID}/events", handler.ListPublic)
	mux.HandleFunc("GET /api/v1/competitions/{competitionID}/events/{eventID}", handler.GetPublic)
	mux.Handle("GET /api/v1/admin/competitions/{competitionID}/events",
		protected(competitionevent.PermissionRead, handler.ListAdmin))
	mux.Handle("GET /api/v1/admin/competitions/{competitionID}/events/{eventID}",
		protected(competitionevent.PermissionRead, handler.GetAdmin))
	mux.Handle("POST /api/v1/competitions/{competitionID}/events",
		protected(competitionevent.PermissionCreate, handler.Create))
	mux.Handle("PATCH /api/v1/competitions/{competitionID}/events/{eventID}",
		protected(competitionevent.PermissionUpdate, handler.Update))
	mux.Handle("DELETE /api/v1/competitions/{competitionID}/events/{eventID}",
		protected(competitionevent.PermissionDelete, handler.Delete))
}
