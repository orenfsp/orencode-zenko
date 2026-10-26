package audit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	auditerrs "github.com/orenfsp/orencode-zenko.git/internal/audit/errs"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"github.com/google/uuid"
)

type repositoryStub struct {
	events []Event
	err    error
	limit  int32
	offset int32
}

func (r *repositoryStub) Write(_ context.Context, _ EventCreate) (Event, error) {
	return Event{}, nil
}

func (r *repositoryStub) List(_ context.Context, limit, offset int32) ([]Event, error) {
	r.limit, r.offset = limit, offset
	return r.events, r.err
}

func TestListMapsEventsAndPassesPagination(t *testing.T) {
	userID, callID := uuid.New(), uuid.New()
	when := time.Date(2026, time.September, 21, 12, 0, 0, 0, time.UTC)
	repo := &repositoryStub{events: []Event{{
		ID:              uuid.New(),
		EventType:       EventHospitalEquipmentStatusChanged,
		ActorKind:       ActorEmployee,
		ActorUserID:     &userID,
		EntityType:      "hospital_equipment",
		EntityID:        uuid.New(),
		EmergencyCallID: &callID,
		Payload:         json.RawMessage(`{"previous_status":"available","new_status":"busy"}`),
		OccurredAt:      when,
	}}}

	items, err := NewService(repo).List(context.Background(), 25, 50)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.limit != 25 || repo.offset != 50 {
		t.Fatalf("pagination = (%d, %d), want (25, 50)", repo.limit, repo.offset)
	}
	if len(items) != 1 || items[0].ActorUserID == nil || *items[0].ActorUserID != userID.String() {
		t.Fatalf("items = %+v; actor user id was not mapped", items)
	}
	if items[0].OccurredAt != "2026-09-21T12:00:00Z" {
		t.Fatalf("occurred_at = %q", items[0].OccurredAt)
	}
}

func TestListRejectsInvalidPaginationBeforeRepository(t *testing.T) {
	repo := &repositoryStub{}
	_, err := NewService(repo).List(context.Background(), 101, 0)
	if !errors.Is(err, auditerrs.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
	if repo.limit != 0 || repo.offset != 0 {
		t.Fatalf("repository must not be called; pagination = (%d, %d)", repo.limit, repo.offset)
	}
}

func TestNewDeviceEventSetsOnlyDeviceActor(t *testing.T) {
	deviceID := uuid.New()
	event, err := NewDeviceEvent(
		EventDestinationResolved,
		"emergency_call_destination",
		uuid.New(),
		deviceID,
		nil,
		map[string]string{"status": "selected"},
	)
	if err != nil {
		t.Fatalf("NewDeviceEvent() error = %v", err)
	}
	if event.ActorKind != ActorDevice || event.ActorDeviceID == nil || *event.ActorDeviceID != deviceID || event.ActorUserID != nil {
		t.Fatalf("event actor = %+v", event)
	}
	if string(event.Payload) != `{"status":"selected"}` {
		t.Fatalf("payload = %s", event.Payload)
	}
}
