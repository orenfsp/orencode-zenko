package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	auditerrs "github.com/orenfsp/orencode-zenko.git/internal/audit/errs"
)

type repositoryStub struct {
	events  []Event
	written []EventCreate
	err     error
	limit   int32
	offset  int32
}

func (r *repositoryStub) Write(_ context.Context, event EventCreate) error {
	r.written = append(r.written, event)
	return nil
}

func (r *repositoryStub) List(_ context.Context, limit, offset int32) ([]Event, error) {
	r.limit = limit
	r.offset = offset
	return r.events, r.err
}

func TestListMapsEventsAndPassesPagination(t *testing.T) {
	entityID := int64(42)
	actorID := int64(1)
	occurredAt := time.Date(2026, time.October, 27, 12, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	repo := &repositoryStub{events: []Event{{
		ID:          7,
		EventType:   EventRoleCreated,
		ActorUserID: &actorID,
		EntityType:  "role",
		EntityID:    &entityID,
		Method:      "POST",
		Route:       "/api/v1/admin/roles",
		Status:      201,
		OccurredAt:  occurredAt,
	}}}

	items, err := NewService(repo).List(context.Background(), 20, 40)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.limit != 20 || repo.offset != 40 {
		t.Fatalf("pagination = %d/%d, want 20/40", repo.limit, repo.offset)
	}
	if len(items) != 1 || items[0].OccurredAt != "2026-10-27T09:30:00Z" {
		t.Fatalf("items = %#v", items)
	}
}

func TestListRejectsInvalidPagination(t *testing.T) {
	service := NewService(&repositoryStub{})

	_, err := service.List(context.Background(), 0, 0)
	if !errors.Is(err, auditerrs.ErrInvalidRequest) {
		t.Fatalf("List() error = %v, want ErrInvalidRequest", err)
	}
}
