package integration

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type identityContextKey struct{}

type Middleware struct {
	tokens     *token.Manager
	repository Repository
}

func NewMiddleware(tokens *token.Manager, repository Repository) *Middleware {
	return &Middleware{tokens: tokens, repository: repository}
}

func (m *Middleware) Require(permission string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "integration access token required")
			return
		}
		claims, err := m.tokens.Parse(parts[1])
		if err != nil || claims.Type != "access" || claims.PrincipalType != "integration" || claims.IntegrationClientID <= 0 {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid integration access token")
			return
		}
		client, err := m.repository.GetClientByID(r.Context(), claims.IntegrationClientID)
		if errors.Is(err, ErrUnauthorized) {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "integration client is inactive")
			return
		}
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			return
		}
		allowed, err := m.repository.HasPermission(r.Context(), client.ID, permission)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			return
		}
		if !allowed {
			httpx.WriteError(w, http.StatusForbidden, "forbidden", "integration permission denied")
			return
		}
		identity := Identity{ClientID: client.ID, OwnerUserID: client.OwnerUserID, ClientKey: client.ClientKey}
		ctx := context.WithValue(r.Context(), identityContextKey{}, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(Identity)
	return identity, ok
}
