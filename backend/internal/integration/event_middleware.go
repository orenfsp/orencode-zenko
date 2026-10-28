package integration

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PlatformEventMiddleware struct{ publisher Publisher }

func NewPlatformEventMiddleware(publisher Publisher) *PlatformEventMiddleware {
	return &PlatformEventMiddleware{publisher: publisher}
}

func (m *PlatformEventMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		eventType, data, ok := platformEventForRequest(r.Method, r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		writer := &eventStatusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(writer, r)
		if writer.status < 200 || writer.status >= 300 {
			return
		}
		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
		defer cancel()
		if err := m.publisher.Publish(ctx, eventType, data); err != nil {
			logWebhookError("enqueue platform webhook", err)
		}
	})
}

func platformEventForRequest(method, path string) (string, map[string]any, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if method == http.MethodPost && len(parts) == 6 && matchPrefix(parts, "api", "v1", "admin", "competitions") && parts[5] == "publish" {
		id, ok := positiveID(parts[4])
		return EventCompetitionPublished, map[string]any{"competition_id": id}, ok
	}
	if method == http.MethodPost && len(parts) == 6 && matchPrefix(parts, "api", "v1", "admin", "competition-applications") && parts[5] == "approve" {
		id, ok := positiveID(parts[4])
		return EventApplicationApproved, map[string]any{"application_id": id}, ok
	}
	if method == http.MethodPost && len(parts) == 9 && matchPrefix(parts, "api", "v1", "admin", "competitions") && parts[5] == "events" && parts[7] == "results" && parts[8] == "approve" {
		competitionID, competitionOK := positiveID(parts[4])
		eventID, eventOK := positiveID(parts[6])
		return EventProtocolPublished, map[string]any{"competition_id": competitionID, "event_id": eventID}, competitionOK && eventOK
	}
	if method == http.MethodPut && len(parts) == 8 && matchPrefix(parts, "api", "v1", "admin", "competitions") && parts[5] == "events" && parts[7] == "protocol" {
		competitionID, competitionOK := positiveID(parts[4])
		eventID, eventOK := positiveID(parts[6])
		return EventProtocolPublished, map[string]any{"competition_id": competitionID, "event_id": eventID}, competitionOK && eventOK
	}
	return "", nil, false
}

func matchPrefix(parts []string, expected ...string) bool {
	if len(parts) < len(expected) {
		return false
	}
	for index, value := range expected {
		if parts[index] != value {
			return false
		}
	}
	return true
}

func positiveID(raw string) (int64, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	return id, err == nil && id > 0
}

type eventStatusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *eventStatusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *eventStatusWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
