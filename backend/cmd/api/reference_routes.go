package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/reference"
)

func registerReferenceRoutes(mux *http.ServeMux, handler *reference.Handler) {
	mux.HandleFunc("GET /api/v1/disciplines", handler.ListDisciplines)
	mux.HandleFunc("GET /api/v1/competition-types", handler.ListCompetitionTypes)
	mux.HandleFunc("GET /api/v1/competition-levels", handler.ListCompetitionLevels)
	mux.HandleFunc("GET /api/v1/age-groups", handler.ListAgeGroups)
}
