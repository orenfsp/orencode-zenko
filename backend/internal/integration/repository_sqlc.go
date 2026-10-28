package integration

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
	pool    interface {
		BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	}
}

func NewRepository(queries *db.Queries, pool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) Repository {
	return &sqlRepository{queries: queries, pool: pool}
}

func (r *sqlRepository) GetClientByKey(ctx context.Context, clientKey string) (Client, error) {
	row, err := r.queries.GetIntegrationClientByKey(ctx, clientKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return Client{}, ErrInvalidCredentials
	}
	if err != nil {
		return Client{}, err
	}
	permissions, err := r.queries.ListIntegrationClientPermissions(ctx, row.ID)
	if err != nil {
		return Client{}, err
	}
	return Client{ID: row.ID, ClientKey: row.ClientKey, Name: row.Name, SecretHash: row.SecretHash,
		OwnerUserID: row.OwnerUserID, Permissions: permissions}, nil
}

func (r *sqlRepository) GetClientByID(ctx context.Context, clientID int64) (Client, error) {
	row, err := r.queries.GetIntegrationClientByID(ctx, clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Client{}, ErrUnauthorized
	}
	if err != nil {
		return Client{}, err
	}
	return Client{ID: row.ID, ClientKey: row.ClientKey, Name: row.Name, OwnerUserID: row.OwnerUserID}, nil
}

func (r *sqlRepository) HasPermission(ctx context.Context, clientID int64, permission string) (bool, error) {
	return r.queries.IntegrationClientHasPermission(ctx, db.IntegrationClientHasPermissionParams{
		ClientID: clientID, PermissionCode: permission,
	})
}

func (r *sqlRepository) CreateSubscription(ctx context.Context, clientID int64, request CreateWebhookRequest) (WebhookSubscription, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WebhookSubscription{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	row, err := queries.CreateWebhookSubscription(ctx, db.CreateWebhookSubscriptionParams{
		ClientID: clientID, Url: request.URL, SigningSecret: request.SigningSecret,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return WebhookSubscription{}, ErrConflict
		}
		return WebhookSubscription{}, err
	}
	for _, eventType := range request.Events {
		if err := queries.AddWebhookSubscriptionEvent(ctx, db.AddWebhookSubscriptionEventParams{
			SubscriptionID: row.ID, EventType: eventType,
		}); err != nil {
			return WebhookSubscription{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return WebhookSubscription{}, err
	}
	return WebhookSubscription{ID: row.ID, URL: row.Url, Events: request.Events,
		IsActive: row.IsActive, CreatedAt: row.CreatedAt.Time}, nil
}

func (r *sqlRepository) ListSubscriptions(ctx context.Context, clientID int64) ([]WebhookSubscription, error) {
	rows, err := r.queries.ListWebhookSubscriptions(ctx, clientID)
	if err != nil {
		return nil, err
	}
	result := make([]WebhookSubscription, 0, len(rows))
	for _, row := range rows {
		events, err := r.queries.ListWebhookSubscriptionEvents(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, WebhookSubscription{ID: row.ID, URL: row.Url, Events: events,
			IsActive: row.IsActive, CreatedAt: row.CreatedAt.Time})
	}
	return result, nil
}

func (r *sqlRepository) DeleteSubscription(ctx context.Context, clientID, subscriptionID int64) error {
	rows, err := r.queries.DeleteWebhookSubscription(ctx, db.DeleteWebhookSubscriptionParams{
		SubscriptionID: subscriptionID, ClientID: clientID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlRepository) Enqueue(ctx context.Context, eventType string, payload json.RawMessage) error {
	return r.queries.EnqueueWebhookDeliveries(ctx, db.EnqueueWebhookDeliveriesParams{EventType: eventType, Payload: payload})
}

func (r *sqlRepository) ClaimDeliveries(ctx context.Context, batchSize int32) ([]Delivery, error) {
	rows, err := r.queries.ClaimWebhookDeliveries(ctx, batchSize)
	if err != nil {
		return nil, err
	}
	result := make([]Delivery, 0, len(rows))
	for _, row := range rows {
		result = append(result, Delivery{ID: row.ID, EventType: row.EventType, Payload: row.Payload,
			Attempts: row.Attempts, URL: row.Url, SigningSecret: row.SigningSecret, CreatedAt: row.CreatedAt.Time})
	}
	return result, nil
}

func (r *sqlRepository) MarkDeliverySent(ctx context.Context, deliveryID int64, responseStatus int32) error {
	return r.queries.MarkWebhookDeliverySent(ctx, db.MarkWebhookDeliverySentParams{
		DeliveryID: deliveryID, ResponseStatus: pgtype.Int4{Int32: responseStatus, Valid: true},
	})
}

func (r *sqlRepository) MarkDeliveryFailed(
	ctx context.Context, deliveryID int64, responseStatus *int32, message, status string, nextAttempt time.Time,
) error {
	return r.queries.RetryWebhookDelivery(ctx, db.RetryWebhookDeliveryParams{
		DeliveryID: deliveryID, Status: status,
		ResponseStatus: optionalInt4(responseStatus), LastError: pgtype.Text{String: message, Valid: message != ""},
		NextAttemptAt: pgtype.Timestamptz{Time: nextAttempt, Valid: true},
	})
}

func optionalInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}
