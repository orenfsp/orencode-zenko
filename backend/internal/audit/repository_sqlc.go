package audit

import (
	"context"
	"encoding/json"

	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: value, Valid: true}
}

func optionalPGUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*value)
}

func optionalUUID(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id := uuid.UUID(value.Bytes)
	return &id
}

func toEvent(row db.AuditLog) Event {
	return Event{
		ID:              uuid.UUID(row.ID.Bytes),
		EventType:       row.EventType,
		ActorKind:       row.ActorKind,
		ActorUserID:     optionalUUID(row.ActorUserID),
		ActorDeviceID:   optionalUUID(row.ActorDeviceID),
		EntityType:      row.EntityType,
		EntityID:        uuid.UUID(row.EntityID.Bytes),
		EmergencyCallID: optionalUUID(row.EmergencyCallID),
		Payload:         json.RawMessage(row.Payload),
		OccurredAt:      row.OccurredAt.Time,
	}
}

// WriteEvent persists an audit event through the supplied query executor.
// Other repositories use it with queries bound to their transaction, so a
// business change and its audit record commit or roll back together.
func WriteEvent(ctx context.Context, queries *db.Queries, event EventCreate) (Event, error) {
	row, err := queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		EventType:       event.EventType,
		ActorKind:       event.ActorKind,
		ActorUserID:     optionalPGUUID(event.ActorUserID),
		ActorDeviceID:   optionalPGUUID(event.ActorDeviceID),
		EntityType:      event.EntityType,
		EntityID:        pgUUID(event.EntityID),
		EmergencyCallID: optionalPGUUID(event.EmergencyCallID),
		Payload:         event.Payload,
	})
	if err != nil {
		return Event{}, err
	}
	return toEvent(row), nil
}

func (r *sqlRepository) Write(ctx context.Context, event EventCreate) (Event, error) {
	return WriteEvent(ctx, r.queries, event)
}

func (r *sqlRepository) List(ctx context.Context, limit, offset int32) ([]Event, error) {
	rows, err := r.queries.ListAuditLogs(ctx, db.ListAuditLogsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	result := make([]Event, 0, len(rows))
	for _, row := range rows {
		result = append(result, toEvent(row))
	}
	return result, nil
}
