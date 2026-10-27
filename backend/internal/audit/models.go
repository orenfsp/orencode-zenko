package audit

import "time"

const (
	EventRoleCreated      = "role.created"
	EventUserRoleAssigned = "user.role.assigned"
	EventProfileUpdated   = "user.profile.updated"
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
