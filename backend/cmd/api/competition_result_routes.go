package main

import (
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/internal/result"
)

func registerCompetitionResultRoutes(
	mux *http.ServeMux, handler *result.Handler, protected protectedHandler,
) {
	base := "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results"
	mux.Handle("GET "+base, protected(result.PermissionRead, handler.Get))
	mux.Handle("PUT "+base+"/settings", protected(result.PermissionManage, handler.Configure))
	mux.Handle("PUT "+base, protected(result.PermissionManage, handler.Save))
	mux.Handle("POST "+base+"/calculate", protected(result.PermissionManage, handler.Calculate))
	mux.Handle("POST "+base+"/approve", protected(result.PermissionApprove, handler.Approve))
	mux.HandleFunc(
		"GET /api/v1/competitions/{competitionID}/events/{eventID}/protocol",
		handler.DownloadProtocol,
	)
}
