package audit

import (
	"context"
	"encoding/json"
	"time"

	auditerrs "github.com/orenfsp/orencode-zenko.git/internal/audit/errs"
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func marshalPayload(value any) (json.RawMessage, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func NewEmployeeEvent(eventType, entityType string, entityID, userID uuid.UUID, emergencyCallID *uuid.UUID, payload any) (EventCreate, error) {
	data, err := marshalPayload(payload)
	if err != nil {
		return EventCreate{}, err
	}
	return EventCreate{
		EventType: eventType, ActorKind: ActorEmployee, ActorUserID: &userID,
		EntityType: entityType, EntityID: entityID, EmergencyCallID: emergencyCallID, Payload: data,
	}, nil
}

func NewDeviceEvent(eventType, entityType string, entityID, deviceID uuid.UUID, emergencyCallID *uuid.UUID, payload any) (EventCreate, error) {
	data, err := marshalPayload(payload)
	if err != nil {
		return EventCreate{}, err
	}
	return EventCreate{
		EventType: eventType, ActorKind: ActorDevice, ActorDeviceID: &deviceID,
		EntityType: entityType, EntityID: entityID, EmergencyCallID: emergencyCallID, Payload: data,
	}, nil
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]ResponseDTO, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, auditerrs.ErrInvalidRequest
	}
	values, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]ResponseDTO, 0, len(values))
	for _, value := range values {
		response := ResponseDTO{
			ID: value.ID.String(), EventType: value.EventType, ActorKind: value.ActorKind,
			EntityType: value.EntityType, EntityID: value.EntityID.String(),
			Payload: value.Payload, OccurredAt: value.OccurredAt.UTC().Format(time.RFC3339),
		}
		if value.ActorUserID != nil {
			actorUserID := value.ActorUserID.String()
			response.ActorUserID = &actorUserID
		}
		if value.ActorDeviceID != nil {
			actorDeviceID := value.ActorDeviceID.String()
			response.ActorDeviceID = &actorDeviceID
		}
		if value.EmergencyCallID != nil {
			emergencyCallID := value.EmergencyCallID.String()
			response.EmergencyCallID = &emergencyCallID
		}
		result = append(result, response)
	}
	return result, nil
}
