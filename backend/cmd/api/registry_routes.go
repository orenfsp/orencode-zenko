package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/registry"
)

func registerRegistryRoutes(mux *http.ServeMux, handler *registry.Handler, protected protectedHandler) {
	mux.Handle("GET /api/v1/admin/registry/users", protected(registry.PermissionRead, handler.ListUsers))
	mux.Handle("GET /api/v1/admin/registry/organizations", protected(registry.PermissionRead, handler.ListOrganizations))
}
