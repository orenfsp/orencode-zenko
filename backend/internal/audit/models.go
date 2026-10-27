package audit

import "time"

const (
	EventRoleCreated      = "role.created"
	EventUserRoleAssigned = "user.role.assigned"
	EventProfileUpdated   = "user.profile.updated"
	EventResultConfigured = "competition_result.configured"
	EventResultUpdated    = "competition_result.updated"
	EventResultCalculated = "competition_result.calculated"
	EventResultApproved   = "competition_result.approved"
)

type EventCreate struct {
	EventType   string
	ActorUserID int64
	EntityType  string
	EntityID    *int64
	Method      string
	Route       string
	Status      int16
}

type Event struct {
	ID          int64
	EventType   string
	ActorUserID int64
	EntityType  string
	EntityID    *int64
	Method      string
	Route       string
	Status      int16
	OccurredAt  time.Time
}
