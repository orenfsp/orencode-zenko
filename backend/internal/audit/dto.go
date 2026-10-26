package audit

import "encoding/json"

type ResponseDTO struct {
	ID              string          `json:"id"`
	EventType       string          `json:"event_type"`
	ActorKind       string          `json:"actor_kind"`
	ActorUserID     *string         `json:"actor_user_id"`
	ActorDeviceID   *string         `json:"actor_device_id"`
	EntityType      string          `json:"entity_type"`
	EntityID        string          `json:"entity_id"`
	EmergencyCallID *string         `json:"emergency_call_id"`
	Payload         json.RawMessage `json:"payload"`
	OccurredAt      string          `json:"occurred_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
