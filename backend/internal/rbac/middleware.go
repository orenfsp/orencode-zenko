package rbac

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type identityContextKey struct{}

type Middleware struct {
	tokens      *token.Manager
	permissions Repository
}

func NewMiddleware(tokens *token.Manager, permissions Repository) *Middleware {
	return &Middleware{tokens: tokens, permissions: permissions}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeMiddlewareError(w, http.StatusUnauthorized, "требуется access-токен")
			return
		}
		claims, err := m.tokens.Parse(rawToken)
		if err != nil || claims.Type != "access" || claims.PrincipalType != "user" || claims.UserID <= 0 {
			writeMiddlewareError(w, http.StatusUnauthorized, "недействительный access-токен")
			return
		}
		identity := Identity{UserID: claims.UserID, RoleID: claims.RoleID}
		ctx := context.WithValue(r.Context(), identityContextKey{}, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) Require(permissionCode string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok {
				writeMiddlewareError(w, http.StatusUnauthorized, "требуется авторизация")
				return
			}
			allowed, err := m.permissions.HasPermission(r.Context(), identity.UserID, permissionCode)
			if err != nil {
				writeMiddlewareError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
				return
			}
			if !allowed {
				writeMiddlewareError(w, http.StatusForbidden, "недостаточно прав")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(Identity)
	return identity, ok
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func writeMiddlewareError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
