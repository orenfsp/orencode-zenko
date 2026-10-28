package integration

import (
	"encoding/json"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/result"
)

const (
	PermissionResultsImport  = "integration.results.import"
	PermissionWebhooksManage = "integration.webhooks.manage"

	EventCompetitionPublished = "competition.published"
	EventApplicationApproved  = "application.approved"
	EventProtocolPublished    = "protocol.published"
)

var AllowedPermissions = map[string]struct{}{
	PermissionResultsImport:  {},
	PermissionWebhooksManage: {},
}

var AllowedWebhookEvents = map[string]struct{}{
	EventCompetitionPublished: {},
	EventApplicationApproved:  {},
	EventProtocolPublished:    {},
}

type Client struct {
	ID          int64
	ClientKey   string
	Name        string
	SecretHash  string
	OwnerUserID int64
	Permissions []string
}

type Identity struct {
	ClientID    int64
	OwnerUserID int64
	ClientKey   string
}

type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	Permissions []string `json:"permissions"`
}

type CreateWebhookRequest struct {
	URL           string   `json:"url"`
	SigningSecret string   `json:"signing_secret"`
	Events        []string `json:"events"`
}

type WebhookSubscription struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type ImportEntry struct {
	EntryID      int64                `json:"entry_id"`
	ResultStatus string               `json:"result_status"`
	Tasks        []result.TaskRequest `json:"tasks,omitempty"`
	Place        *int32               `json:"place,omitempty"`
	CustomValues map[string]any       `json:"custom_values,omitempty"`
}

type ImportResultsRequest struct {
	Revision int64         `json:"revision,omitempty"`
	Entries  []ImportEntry `json:"entries"`
}

type PlatformEvent struct {
	Type       string `json:"type"`
	OccurredAt string `json:"occurred_at"`
	Data       any    `json:"data"`
}

type Delivery struct {
	ID            int64
	EventType     string
	Payload       json.RawMessage
	Attempts      int32
	URL           string
	SigningSecret string
	CreatedAt     time.Time
}
