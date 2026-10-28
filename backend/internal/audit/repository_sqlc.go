package audit

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) Write(ctx context.Context, event EventCreate) error {
	return r.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		EventType:   event.EventType,
		ActorUserID: optionalInt8(event.ActorUserID),
		EntityType:  event.EntityType,
		EntityID:    optionalInt8(event.EntityID),
		Method:      event.Method,
		Route:       event.Route,
		Status:      event.Status,
	})
}

func (r *sqlRepository) List(ctx context.Context, limit, offset int32) ([]Event, error) {
	rows, err := r.queries.ListAuditLogs(ctx, db.ListAuditLogsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	result := make([]Event, 0, len(rows))
	for _, row := range rows {
		result = append(result, Event{
			ID:          row.ID,
			EventType:   row.EventType,
			ActorUserID: optionalInt64(row.ActorUserID),
			EntityType:  row.EntityType,
			EntityID:    optionalInt64(row.EntityID),
			Method:      row.Method,
			Route:       row.Route,
			Status:      row.Status,
			OccurredAt:  row.OccurredAt.Time,
		})
	}
	return result, nil
}

func optionalInt8(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func optionalInt64(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}
