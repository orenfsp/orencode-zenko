package integration

import (
	"net/http"
	"testing"
)

func TestPlatformEventForRequest(t *testing.T) {
	tests := []struct {
		method, path, eventType string
	}{
		{http.MethodPost, "/api/v1/admin/competitions/12/publish", EventCompetitionPublished},
		{http.MethodPost, "/api/v1/admin/competition-applications/13/approve", EventApplicationApproved},
		{http.MethodPost, "/api/v1/admin/competitions/12/events/14/results/approve", EventProtocolPublished},
		{http.MethodPut, "/api/v1/admin/competitions/12/events/14/protocol", EventProtocolPublished},
	}
	for _, tt := range tests {
		eventType, _, ok := platformEventForRequest(tt.method, tt.path)
		if !ok || eventType != tt.eventType {
			t.Errorf("platformEventForRequest(%q, %q) = %q, %v", tt.method, tt.path, eventType, ok)
		}
	}
}
