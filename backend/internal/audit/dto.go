package audit

type ResponseDTO struct {
	ID          int64  `json:"id"`
	EventType   string `json:"event_type"`
	ActorUserID int64  `json:"actor_user_id"`
	EntityType  string `json:"entity_type"`
	EntityID    *int64 `json:"entity_id,omitempty"`
	Method      string `json:"method"`
	Route       string `json:"route"`
	Status      int16  `json:"status"`
	OccurredAt  string `json:"occurred_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
