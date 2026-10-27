package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/registry"
)

func TestRegisterRegistryRoutes(t *testing.T) {
	mux := http.NewServeMux()
	permissions := make([]string, 0, 2)
	handler := registry.NewHandler(registry.NewService(&registryRouteRepository{}))
	registerRegistryRoutes(mux, handler, func(permission string, handler http.HandlerFunc) http.Handler {
		permissions = append(permissions, permission)
		return handler
	})
	for _, path := range []string{"/api/v1/admin/registry/users", "/api/v1/admin/registry/organizations"} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, response.Code)
		}
	}
	if len(permissions) != 2 || permissions[0] != registry.PermissionRead || permissions[1] != registry.PermissionRead {
		t.Fatalf("permissions=%v", permissions)
	}
}

type registryRouteRepository struct{}

func (*registryRouteRepository) ListUsers(_ context.Context, filter registry.UserFilter) (registry.Page[registry.User], error) {
	return registry.Page[registry.User]{Items: []registry.User{}, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (*registryRouteRepository) ListOrganizations(_ context.Context, filter registry.OrganizationFilter) (registry.Page[registry.Organization], error) {
	return registry.Page[registry.Organization]{Items: []registry.Organization{}, Limit: filter.Limit, Offset: filter.Offset}, nil
}
