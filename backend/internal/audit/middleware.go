package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

		var refreshToken string
		if rule.eventType == EventLogout {
			refreshToken = readRefreshToken(r)
		}

		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(writer, r)

		actorID := m.accessTokenActorID(r)
		if actorID == nil && rule.eventType == EventLogin && writer.status >= 200 && writer.status < 300 {
			actorID = m.loginResponseActorID(writer.body.Bytes())
		}
		if actorID == nil && rule.eventType == EventLogout && refreshToken != "" {
			actorID = m.tokenActorID(refreshToken, "refresh")
		}
		if rule.entityIDFromActor {
			rule.entityID = actorID
		}
		if rule.entityIDFromResponse && writer.status >= 200 && writer.status < 300 {
			rule.entityID = responseEntityID(writer.body.Bytes())
		}

		err := m.repo.Write(context.WithoutCancel(r.Context()), EventCreate{
			EventType: rule.eventType, ActorUserID: actorID,
			EntityType: rule.entityType, EntityID: rule.entityID,
			Method: r.Method, Route: rule.route, Status: int16(writer.status),
		})
		if err != nil && logger.Log != nil {
			logger.Log.Error("write audit log failed", zap.Error(err))
		}
	})
}

func (m *Middleware) accessTokenActorID(r *http.Request) *int64 {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil
	}
	return m.tokenActorID(parts[1], "access")
}

func (m *Middleware) tokenActorID(rawToken, tokenType string) *int64 {
	claims, err := m.tokens.Parse(rawToken)
	if err != nil || claims.Type != tokenType || claims.PrincipalType != "user" || claims.UserID <= 0 {
		return nil
	}
	actorID := claims.UserID
	return &actorID
}

func (m *Middleware) loginResponseActorID(body []byte) *int64 {
	var response struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &response) != nil || response.Data.AccessToken == "" {
		return nil
	}
	return m.tokenActorID(response.Data.AccessToken, "access")
}

func readRefreshToken(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	if json.Unmarshal(body, &request) != nil {
		return ""
	}
	return request.RefreshToken
}

func responseEntityID(body []byte) *int64 {
	var response struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &response) != nil || response.Data.ID <= 0 {
		return nil
	}
	return &response.Data.ID
}

type routeRule struct {
	eventType            string
	entityType           string
	entityID             *int64
	entityIDFromActor    bool
	entityIDFromResponse bool
	route                string
}

func auditedRoute(method, path string) (routeRule, bool) {
	segments := pathSegments(path)

	switch {
	case method == http.MethodPost && path == "/api/v1/auth/login":
		return routeRule{eventType: EventLogin, entityType: "user", entityIDFromActor: true, route: path}, true
	case method == http.MethodPost && path == "/api/v1/auth/logout":
		return routeRule{eventType: EventLogout, entityType: "user", entityIDFromActor: true, route: path}, true
	case method == http.MethodPost && path == "/api/v1/admin/roles":
		return routeRule{eventType: EventRoleCreated, entityType: "role", entityIDFromResponse: true, route: path}, true
	case method == http.MethodPatch && path == "/api/v1/profile":
		return routeRule{eventType: EventProfileUpdated, entityType: "user", entityIDFromActor: true, route: path}, true
	case method == http.MethodPost && len(segments) == 6 && segments[0] == "api" && segments[1] == "v1" && segments[2] == "admin" && segments[3] == "users" && segments[5] == "roles":
		return ruleWithPathID(EventUserRoleAssigned, "user", segments[4], "/api/v1/admin/users/{userID}/roles")
	case method == http.MethodPost && path == "/api/v1/competitions":
		return routeRule{eventType: EventCompetitionCreated, entityType: "competition", entityIDFromResponse: true, route: path}, true
	case method == http.MethodPatch && len(segments) == 4 && segments[0] == "api" && segments[1] == "v1" && segments[2] == "competitions":
		return ruleWithPathID(EventCompetitionUpdated, "competition", segments[3], "/api/v1/competitions/{id}")
	case method == http.MethodPost && len(segments) == 6 && segments[0] == "api" && segments[1] == "v1" && segments[2] == "admin" && segments[3] == "competitions":
		eventType, ok := competitionStatusEvent(segments[5])
		if !ok {
			return routeRule{}, false
		}
		return ruleWithPathID(eventType, "competition", segments[4], "/api/v1/admin/competitions/{id}/"+segments[5])
	case method == http.MethodPost && len(segments) == 5 && isCompetitionEventPath(segments):
		return routeRule{eventType: EventCompetitionEventCreated, entityType: "competition_event", entityIDFromResponse: true, route: "/api/v1/competitions/{competitionID}/events"}, true
	case (method == http.MethodPatch || method == http.MethodDelete) && len(segments) == 6 && isCompetitionEventPath(segments):
		eventType := EventCompetitionEventUpdated
		if method == http.MethodDelete {
			eventType = EventCompetitionEventDeleted
		}
		return ruleWithPathID(eventType, "competition_event", segments[5], "/api/v1/competitions/{competitionID}/events/{eventID}")
	case method == http.MethodPost && len(segments) == 5 && isCompetitionJudgePath(segments):
		return ruleWithPathID(EventCompetitionJudgeAssigned, "competition", segments[3], "/api/v1/competitions/{id}/judges")
	case method == http.MethodDelete && len(segments) == 6 && isCompetitionJudgePath(segments):
		return ruleWithPathID(EventCompetitionJudgeRemoved, "competition_judge", segments[5], "/api/v1/competitions/{id}/judges/{assignmentId}")
	case method == http.MethodPost && len(segments) == 6 && isAdminApplicationPath(segments, "competition-applications"):
		eventType := EventCompetitionApplicationApprove
		if segments[5] == "reject" {
			eventType = EventCompetitionApplicationReject
		} else if segments[5] != "approve" {
			return routeRule{}, false
		}
		return ruleWithPathID(eventType, "competition_application", segments[4], "/api/v1/admin/competition-applications/{id}/"+segments[5])
	case method == http.MethodPatch && len(segments) == 5 && isAdminApplicationPath(segments, "judge-category-applications"):
		return ruleWithPathID(EventJudgeCategoryReviewed, "judge_category_application", segments[4], "/api/v1/admin/judge-category-applications/{id}")
	case method == http.MethodPatch && len(segments) == 5 && isAdminApplicationPath(segments, "rank-applications"):
		return ruleWithPathID(EventRankApplicationReviewed, "rank_application", segments[4], "/api/v1/admin/rank-applications/{applicationID}")
	case method == http.MethodPatch && len(segments) == 5 && isAdminApplicationPath(segments, "anti-doping"):
		return ruleWithPathID(EventAntiDopingReviewed, "anti_doping", segments[4], "/api/v1/admin/anti-doping/{certificateID}")
	case method == http.MethodPatch && len(segments) == 5 && isAdminApplicationPath(segments, "parental-consents"):
		return ruleWithPathID(EventParentalConsentReviewed, "parental_consent", segments[4], "/api/v1/admin/parental-consents/{id}")
	case method == http.MethodPut && strings.HasSuffix(path, "/results/settings"):
		return resultRouteRule(path, EventResultConfigured, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results/settings")
	case method == http.MethodPut && strings.HasSuffix(path, "/results"):
		return resultRouteRule(path, EventResultUpdated, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results")
	case method == http.MethodPost && strings.HasSuffix(path, "/results/calculate"):
		return resultRouteRule(path, EventResultCalculated, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results/calculate")
	case method == http.MethodPost && strings.HasSuffix(path, "/results/approve"):
		return resultRouteRule(path, EventResultApproved, "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results/approve")
	case (method == http.MethodPut || method == http.MethodDelete) && len(segments) == 8 && isAdminEventPath(segments) && segments[7] == "protocol":
		eventType := EventProtocolUploaded
		if method == http.MethodDelete {
			eventType = EventProtocolDeleted
		}
		return ruleWithPathID(eventType, "competition_event", segments[6], "/api/v1/admin/competitions/{competitionID}/events/{eventID}/protocol")
	case method == http.MethodPut && path == "/api/v1/admin/rating/selections":
		return routeRule{eventType: EventRatingSelectionSaved, entityType: "rating_selection", entityIDFromResponse: true, route: path}, true
	case method == http.MethodPost && len(segments) == 7 && segments[0] == "api" && segments[1] == "v1" && segments[2] == "admin" && segments[3] == "rating" && segments[4] == "selections" && segments[6] == "lock":
		return ruleWithPathID(EventRatingSelectionLocked, "rating_selection", segments[5], "/api/v1/admin/rating/selections/{selectionID}/lock")
	default:
		return routeRule{}, false
	}
}

func competitionStatusEvent(action string) (string, bool) {
	switch action {
	case "submit":
		return EventCompetitionSubmitted, true
	case "publish":
		return EventCompetitionPublished, true
	case "cancel":
		return EventCompetitionCancelled, true
	case "return":
		return EventCompetitionReturned, true
	default:
		return "", false
	}
}

func ruleWithPathID(eventType, entityType, rawID, route string) (routeRule, bool) {
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return routeRule{}, false
	}
	return routeRule{eventType: eventType, entityType: entityType, entityID: &id, route: route}, true
}

func pathSegments(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

func isCompetitionEventPath(parts []string) bool {
	return len(parts) >= 5 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "competitions" && parts[4] == "events"
}

func isCompetitionJudgePath(parts []string) bool {
	return len(parts) >= 5 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "competitions" && parts[4] == "judges"
}

func isAdminApplicationPath(parts []string, resource string) bool {
	return len(parts) >= 5 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "admin" && parts[3] == resource
}

func isAdminEventPath(parts []string) bool {
	return len(parts) >= 8 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "admin" && parts[3] == "competitions" && parts[5] == "events"
}

func resultRouteRule(path, eventType, route string) (routeRule, bool) {
	parts := pathSegments(path)
	if !isAdminEventPath(parts) || len(parts) < 8 || parts[7] != "results" {
		return routeRule{}, false
	}
	return ruleWithPathID(eventType, "competition_event", parts[6], route)
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}
