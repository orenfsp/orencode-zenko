package audit

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
	"go.uber.org/zap"
)

type Middleware struct {
	repo   Repository
	tokens *token.Manager
}

func NewMiddleware(repo Repository, tokens *token.Manager) *Middleware {
	return &Middleware{repo: repo, tokens: tokens}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rule, ok := auditedRoute(r.Method, r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(writer, r)
		if writer.status < 200 || writer.status >= 300 {
			return
		}

		actorID, ok := m.actorID(r)
		if !ok {
			return
		}
		if rule.entityIDFromActor {
			rule.entityID = &actorID
		}

		err := m.repo.Write(context.WithoutCancel(r.Context()), EventCreate{
			EventType: rule.eventType, ActorUserID: actorID,
			EntityType: rule.entityType, EntityID: rule.entityID,
			Method: r.Method, Route: rule.route, Status: int16(writer.status),
		})
		if err != nil {
			logger.Log.Error("write audit log failed", zap.Error(err))
		}
	})
}

func (m *Middleware) actorID(r *http.Request) (int64, bool) {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return 0, false
	}
	claims, err := m.tokens.Parse(parts[1])
	if err != nil || claims.Type != "access" || claims.PrincipalType != "user" || claims.UserID <= 0 {
		return 0, false
	}
	return claims.UserID, true
}

type routeRule struct {
	eventType         string
	entityType        string
	entityID          *int64
	entityIDFromActor bool
	route             string
}

func auditedRoute(method, path string) (routeRule, bool) {
	switch {
	case method == http.MethodPut && strings.HasSuffix(path, "/results/settings"):
		return resultRouteRule(path, EventResultConfigured, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results/settings")
	case method == http.MethodPut && strings.HasSuffix(path, "/results"):
		return resultRouteRule(path, EventResultUpdated, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results")
	case method == http.MethodPost && strings.HasSuffix(path, "/results/calculate"):
		return resultRouteRule(path, EventResultCalculated, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results/calculate")
	case method == http.MethodPost && strings.HasSuffix(path, "/results/approve"):
		return resultRouteRule(path, EventResultApproved, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results/approve")
	case method == http.MethodPost && path == "/api/v1/admin/roles":
		return routeRule{eventType: EventRoleCreated, entityType: "role", route: path}, true
	case method == http.MethodPatch && path == "/api/v1/profile":
		return routeRule{eventType: EventProfileUpdated, entityType: "user", entityIDFromActor: true, route: path}, true
	case method == http.MethodPost && strings.HasPrefix(path, "/api/v1/admin/users/") && strings.HasSuffix(path, "/roles"):
		rawID := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/admin/users/"), "/roles")
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return routeRule{}, false
		}
		return routeRule{
			eventType: EventUserRoleAssigned, entityType: "user", entityID: &id,
			route: "/api/v1/admin/users/{userID}/roles",
		}, true
	default:
		return routeRule{}, false
	}
}

func resultRouteRule(path, eventType, route string) (routeRule, bool) {
	const prefix = "/api/v1/admin/competitions/"
	if !strings.HasPrefix(path, prefix) {
		return routeRule{}, false
	}
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) < 4 || parts[1] != "events" || parts[3] != "results" {
		return routeRule{}, false
	}
	eventID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || eventID <= 0 {
		return routeRule{}, false
	}
	return routeRule{
		eventType: eventType, entityType: "competition_event", entityID: &eventID, route: route,
	}, true
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}
