package integration

import (
	"context"
	"encoding/json"
	"time"
)

type Repository interface {
	GetClientByKey(ctx context.Context, clientKey string) (Client, error)
	GetClientByID(ctx context.Context, clientID int64) (Client, error)
	HasPermission(ctx context.Context, clientID int64, permission string) (bool, error)
	CreateSubscription(ctx context.Context, clientID int64, request CreateWebhookRequest) (WebhookSubscription, error)
	ListSubscriptions(ctx context.Context, clientID int64) ([]WebhookSubscription, error)
	DeleteSubscription(ctx context.Context, clientID, subscriptionID int64) error
	Enqueue(ctx context.Context, eventType string, payload json.RawMessage) error
	ClaimDeliveries(ctx context.Context, batchSize int32) ([]Delivery, error)
	MarkDeliverySent(ctx context.Context, deliveryID int64, responseStatus int32) error
	MarkDeliveryFailed(ctx context.Context, deliveryID int64, responseStatus *int32, message, status string, nextAttempt time.Time) error
}

type Publisher interface {
	Publish(ctx context.Context, eventType string, data any) error
}
