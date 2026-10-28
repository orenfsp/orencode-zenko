package integration

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/result"
	"github.com/orenfsp/orencode-zenko.git/pkg/password"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type ResultService interface {
	Get(context.Context, int64, int64) (result.TableResponse, error)
	Save(context.Context, int64, int64, result.SaveResultsRequest) (result.TableResponse, error)
	SaveManual(context.Context, int64, int64, result.SaveManualResultsRequest) (result.TableResponse, error)
	Calculate(context.Context, int64, int64) (result.TableResponse, error)
	Approve(context.Context, int64, int64, int64) (result.ApprovalResponse, error)
}

type Service struct {
	repository Repository
	tokens     *token.Manager
	results    ResultService
}

func NewService(repository Repository, tokens *token.Manager, results ResultService) *Service {
	return &Service{repository: repository, tokens: tokens, results: results}
}

func (s *Service) Token(ctx context.Context, request TokenRequest) (TokenResponse, error) {
	request.ClientID = strings.TrimSpace(request.ClientID)
	if request.ClientID == "" || request.ClientSecret == "" {
		return TokenResponse{}, ErrInvalidCredentials
	}
	client, err := s.repository.GetClientByKey(ctx, request.ClientID)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return TokenResponse{}, ErrInvalidCredentials
		}
		return TokenResponse{}, err
	}
	if password.Compare(request.ClientSecret, client.SecretHash) != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}
	accessToken, err := s.tokens.GenerateIntegrationAccessToken(token.IntegrationTokenPayload{ClientID: client.ID})
	if err != nil {
		return TokenResponse{}, err
	}
	permissions := append([]string(nil), client.Permissions...)
	sort.Strings(permissions)
	return TokenResponse{AccessToken: accessToken, TokenType: "Bearer",
		ExpiresIn: int64(s.tokens.AccessTTL().Seconds()), Permissions: permissions}, nil
}

func (s *Service) CreateWebhook(ctx context.Context, clientID int64, request CreateWebhookRequest) (WebhookSubscription, error) {
	request.URL = strings.TrimSpace(request.URL)
	request.SigningSecret = strings.TrimSpace(request.SigningSecret)
	request.Events = uniqueStrings(request.Events)
	if clientID <= 0 || !validWebhookURL(request.URL) || len(request.SigningSecret) < 16 || len(request.SigningSecret) > 200 || len(request.Events) == 0 {
		return WebhookSubscription{}, ErrInvalidRequest
	}
	for _, eventType := range request.Events {
		if _, ok := AllowedWebhookEvents[eventType]; !ok {
			return WebhookSubscription{}, ErrInvalidRequest
		}
	}
	return s.repository.CreateSubscription(ctx, clientID, request)
}

func (s *Service) ListWebhooks(ctx context.Context, clientID int64) ([]WebhookSubscription, error) {
	if clientID <= 0 {
		return nil, ErrInvalidRequest
	}
	items, err := s.repository.ListSubscriptions(ctx, clientID)
	if items == nil && err == nil {
		items = make([]WebhookSubscription, 0)
	}
	return items, err
}

func (s *Service) DeleteWebhook(ctx context.Context, clientID, subscriptionID int64) error {
	if clientID <= 0 || subscriptionID <= 0 {
		return ErrInvalidRequest
	}
	return s.repository.DeleteSubscription(ctx, clientID, subscriptionID)
}

func (s *Service) ImportResults(
	ctx context.Context, identity Identity, competitionID, eventID int64, request ImportResultsRequest,
) (result.ApprovalResponse, error) {
	if identity.ClientID <= 0 || identity.OwnerUserID <= 0 || competitionID <= 0 || eventID <= 0 ||
		len(request.Entries) == 0 || len(request.Entries) > 1000 {
		return result.ApprovalResponse{}, ErrInvalidRequest
	}
	table, err := s.results.Get(ctx, competitionID, eventID)
	if err != nil {
		return result.ApprovalResponse{}, err
	}
	if table.Status == result.SheetApproved && table.ApprovedAt != nil && table.ProtocolURL != nil {
		return result.ApprovalResponse{Status: table.Status, ApprovedAt: *table.ApprovedAt, ProtocolURL: *table.ProtocolURL}, nil
	}

	if table.ScoringMode == result.ScoringManual {
		entries := make([]result.ManualEntryRequest, 0, len(request.Entries))
		for _, entry := range request.Entries {
			entries = append(entries, result.ManualEntryRequest{EntryID: entry.EntryID,
				ResultStatus: entry.ResultStatus, Place: entry.Place, CustomValues: entry.CustomValues})
		}
		if _, err = s.results.SaveManual(ctx, competitionID, eventID, result.SaveManualResultsRequest{
			Revision: request.Revision, Entries: entries,
		}); err != nil {
			return result.ApprovalResponse{}, err
		}
	} else if table.ScoringMode == result.ScoringAlgorithmic || table.ScoringMode == "" {
		entries := make([]result.EntryRequest, 0, len(request.Entries))
		for _, entry := range request.Entries {
			entries = append(entries, result.EntryRequest{EntryID: entry.EntryID,
				ResultStatus: entry.ResultStatus, Tasks: entry.Tasks})
		}
		if _, err = s.results.Save(ctx, competitionID, eventID, result.SaveResultsRequest{Entries: entries}); err != nil {
			return result.ApprovalResponse{}, err
		}
	} else {
		return result.ApprovalResponse{}, result.ErrWrongScoringMode
	}
	if _, err = s.results.Calculate(ctx, competitionID, eventID); err != nil {
		return result.ApprovalResponse{}, err
	}
	approval, err := s.results.Approve(ctx, competitionID, eventID, identity.OwnerUserID)
	if err != nil {
		return result.ApprovalResponse{}, err
	}
	_ = s.Publish(context.WithoutCancel(ctx), EventProtocolPublished, map[string]any{
		"competition_id": competitionID, "event_id": eventID, "protocol_url": approval.ProtocolURL,
		"source": "integration", "integration_client_id": identity.ClientID,
	})
	return approval, nil
}

func (s *Service) GetResults(
	ctx context.Context, identity Identity, competitionID, eventID int64,
) (result.TableResponse, error) {
	if identity.ClientID <= 0 || competitionID <= 0 || eventID <= 0 {
		return result.TableResponse{}, ErrInvalidRequest
	}
	return s.results.Get(ctx, competitionID, eventID)
}

func (s *Service) Publish(ctx context.Context, eventType string, data any) error {
	if _, ok := AllowedWebhookEvents[eventType]; !ok {
		return ErrInvalidRequest
	}
	payload, err := json.Marshal(PlatformEvent{Type: eventType, OccurredAt: time.Now().UTC().Format(time.RFC3339Nano), Data: data})
	if err != nil {
		return err
	}
	return s.repository.Enqueue(ctx, eventType, payload)
}

func validWebhookURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	host := parsed.Hostname()
	return parsed.Scheme == "http" && (host == "localhost" || net.ParseIP(host).IsLoopback())
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
