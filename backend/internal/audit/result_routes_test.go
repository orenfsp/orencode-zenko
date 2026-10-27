package audit

import (
	"net/http"
	"testing"
)

func TestResultMutationRoutesAreAudited(t *testing.T) {
	tests := []struct {
		method, path, eventType string
	}{
		{http.MethodPut, "/api/v1/admin/competitions/4/events/7/results/settings", EventResultConfigured},
		{http.MethodPut, "/api/v1/admin/competitions/4/events/7/results", EventResultUpdated},
		{http.MethodPost, "/api/v1/admin/competitions/4/events/7/results/calculate", EventResultCalculated},
		{http.MethodPost, "/api/v1/admin/competitions/4/events/7/results/approve", EventResultApproved},
	}
	for _, test := range tests {
		rule, ok := auditedRoute(test.method, test.path)
		if !ok || rule.eventType != test.eventType || rule.entityID == nil || *rule.entityID != 7 ||
			rule.entityType != "competition_event" {
			t.Fatalf("auditedRoute(%s, %s) = %#v, %v", test.method, test.path, rule, ok)
		}
	}
}
