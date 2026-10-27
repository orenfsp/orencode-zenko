package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/judgecategory"
)

func TestRegisterJudgeCategoryRoutes(t *testing.T) {
	mux := http.NewServeMux()
	permissions := make([]string, 0, 2)
	handler := judgecategory.NewHandler(judgecategory.NewService(&judgeCategoryRouteRepository{}, judgeCategoryRoleStub{}))
	registerJudgeCategoryRoutes(
		mux,
		handler,
		func(handler http.HandlerFunc) http.Handler { return handler },
		func(permission string, handler http.HandlerFunc) http.Handler {
			permissions = append(permissions, permission)
			return handler
		},
	)
	for _, path := range []string{
		"/api/v1/judge-category-applications/mine",
		"/api/v1/judge-category/current",
		"/api/v1/judge/competitions",
		"/api/v1/admin/judge-category-applications",
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code == http.StatusNotFound {
			t.Fatalf("route %s is not registered", path)
		}
	}
	if len(permissions) != 2 || permissions[0] != judgecategory.PermissionReview || permissions[1] != judgecategory.PermissionReview {
		t.Fatalf("permissions=%v", permissions)
	}
}

type judgeCategoryRouteRepository struct{}

func (*judgeCategoryRouteRepository) CategoryExists(context.Context, int64) (bool, error) {
	return true, nil
}
func (*judgeCategoryRouteRepository) Create(context.Context, int64, judgecategory.CreateRequest) (judgecategory.Application, error) {
	return judgecategory.Application{}, nil
}
func (*judgeCategoryRouteRepository) ListMine(context.Context, int64) ([]judgecategory.Application, error) {
	return []judgecategory.Application{}, nil
}
func (*judgeCategoryRouteRepository) ListPending(context.Context, int32, int32) ([]judgecategory.Application, error) {
	return []judgecategory.Application{}, nil
}
func (*judgeCategoryRouteRepository) Review(context.Context, int64, int64, judgecategory.ReviewInput) (judgecategory.Application, error) {
	return judgecategory.Application{}, nil
}
func (*judgeCategoryRouteRepository) GetCurrent(context.Context, int64) (*judgecategory.CurrentCategory, error) {
	return nil, nil
}
func (*judgeCategoryRouteRepository) ListAssignments(context.Context, int64) ([]judgecategory.Assignment, error) {
	return []judgecategory.Assignment{}, nil
}

type judgeCategoryRoleStub struct{}

func (judgeCategoryRoleStub) HasRole(context.Context, int64, string) (bool, error) { return true, nil }
