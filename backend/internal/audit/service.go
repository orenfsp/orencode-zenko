package audit

import (
	"context"
	"time"

	auditerrs "github.com/orenfsp/orencode-zenko.git/internal/audit/errs"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]ResponseDTO, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, auditerrs.ErrInvalidRequest
	}

	events, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]ResponseDTO, 0, len(events))
	for _, event := range events {
		result = append(result, ResponseDTO{
			ID:          event.ID,
			EventType:   event.EventType,
			ActorUserID: event.ActorUserID,
			EntityType:  event.EntityType,
			EntityID:    event.EntityID,
			Method:      event.Method,
			Route:       event.Route,
			Status:      event.Status,
			OccurredAt:  event.OccurredAt.UTC().Format(time.RFC3339),
		})
	}
	return result, nil
}
