package main

import (
	"net/http"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/antidoping"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

func TestRegisterAntiDopingRoutesUsesReviewPermission(t *testing.T) {
	mux := http.NewServeMux()
	handler := antidoping.NewHandler(nil, nil)
	var permission string
	authenticatedCalls := 0
	authenticated := func(next http.HandlerFunc) http.Handler {
		authenticatedCalls++
		return next
	}
	protected := func(got string, next http.HandlerFunc) http.Handler {
		permission = got
		return next
	}

	registerAntiDopingRoutes(mux, handler, authenticated, protected)

	if authenticatedCalls != 2 {
		t.Fatalf("authenticated route count = %d, want 2", authenticatedCalls)
	}
	if permission != rbac.PermissionAntiDopingReview {
		t.Fatalf("review permission = %q, want %q", permission, rbac.PermissionAntiDopingReview)
	}
}
