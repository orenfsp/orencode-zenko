package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/competition"
)

func registerCompetitionJudgeRoutes(
	mux *http.ServeMux, handler *competition.JudgeHandler, protected protectedHandler,
) {
	mux.HandleFunc("GET /api/v1/competitions/{id}/judges", handler.List)
	mux.Handle(
		"GET /api/v1/admin/competitions/{id}/judges",
		protected(competition.PermissionManageJudges, handler.ListAdmin),
	)
	mux.Handle(
		"POST /api/v1/competitions/{id}/judges",
		protected(competition.PermissionManageJudges, handler.Assign),
	)
	mux.Handle(
		"DELETE /api/v1/competitions/{id}/judges/{assignmentId}",
		protected(competition.PermissionManageJudges, handler.Remove),
	)
}
