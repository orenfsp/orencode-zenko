package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/reference"
)

type routeReferenceRepository struct{}

func (routeReferenceRepository) ListDisciplines(context.Context) ([]reference.NamedItem, error) {
	return []reference.NamedItem{}, nil
}
func (routeReferenceRepository) ListCompetitionTypes(context.Context) ([]reference.NamedItem, error) {
	return []reference.NamedItem{}, nil
}
func (routeReferenceRepository) ListCompetitionLevels(context.Context) ([]reference.CompetitionLevel, error) {
	return []reference.CompetitionLevel{}, nil
}
func (routeReferenceRepository) ListAgeGroups(context.Context) ([]reference.AgeGroup, error) {
	return []reference.AgeGroup{}, nil
}

func TestReferenceRoutesAreRegisteredAndPublic(t *testing.T) {
	mux := http.NewServeMux()
	registerReferenceRoutes(mux, reference.NewHandler(reference.NewService(routeReferenceRepository{})))

	for _, path := range []string{
		"/api/v1/disciplines",
		"/api/v1/competition-types",
		"/api/v1/competition-levels",
		"/api/v1/age-groups",
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s without token status = %d", path, response.Code)
		}
	}
}
