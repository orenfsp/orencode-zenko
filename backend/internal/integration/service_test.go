package integration

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/result"
	"github.com/orenfsp/orencode-zenko.git/pkg/password"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
)

type repositoryStub struct {
	client       Client
	permission   bool
	subscription WebhookSubscription
	published    []PlatformEvent
	err          error
}

func (r *repositoryStub) GetClientByKey(context.Context, string) (Client, error) {
	return r.client, r.err
}
func (r *repositoryStub) GetClientByID(context.Context, int64) (Client, error) {
	return r.client, r.err
}
func (r *repositoryStub) HasPermission(context.Context, int64, string) (bool, error) {
	return r.permission, r.err
}
func (r *repositoryStub) CreateSubscription(context.Context, int64, CreateWebhookRequest) (WebhookSubscription, error) {
	return r.subscription, r.err
}
func (r *repositoryStub) ListSubscriptions(context.Context, int64) ([]WebhookSubscription, error) {
	return []WebhookSubscription{r.subscription}, r.err
}
func (r *repositoryStub) DeleteSubscription(context.Context, int64, int64) error { return r.err }
func (r *repositoryStub) Enqueue(_ context.Context, _ string, payload json.RawMessage) error {
	var event PlatformEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	r.published = append(r.published, event)
	return r.err
}
func (r *repositoryStub) ClaimDeliveries(context.Context, int32) ([]Delivery, error) {
	return nil, r.err
}
func (r *repositoryStub) MarkDeliverySent(context.Context, int64, int32) error { return r.err }
func (r *repositoryStub) MarkDeliveryFailed(context.Context, int64, *int32, string, string, time.Time) error {
	return r.err
}

type resultServiceStub struct {
	table        result.TableResponse
	manualSaved  bool
	calculated   bool
	approvedByID int64
	err          error
}

func (s *resultServiceStub) Get(context.Context, int64, int64) (result.TableResponse, error) {
	return s.table, s.err
}
func (s *resultServiceStub) Save(context.Context, int64, int64, result.SaveResultsRequest) (result.TableResponse, error) {
	return s.table, s.err
}
func (s *resultServiceStub) SaveManual(_ context.Context, _, _ int64, _ result.SaveManualResultsRequest) (result.TableResponse, error) {
	s.manualSaved = true
	return s.table, s.err
}
func (s *resultServiceStub) Calculate(context.Context, int64, int64) (result.TableResponse, error) {
	s.calculated = true
	return s.table, s.err
}
func (s *resultServiceStub) Approve(_ context.Context, _, _, actorID int64) (result.ApprovalResponse, error) {
	s.approvedByID = actorID
	return result.ApprovalResponse{Status: result.SheetApproved, ProtocolURL: "/protocol"}, s.err
}

func TestTokenUsesIntegrationPrincipalAndReturnsPermissions(t *testing.T) {
	hash, err := password.Hash("test-secret-123456")
	if err != nil {
		t.Fatal(err)
	}
	repository := &repositoryStub{client: Client{ID: 7, SecretHash: hash,
		Permissions: []string{PermissionWebhooksManage, PermissionResultsImport}}}
	tokens := token.New("jwt-secret", time.Minute, time.Hour)
	response, err := NewService(repository, tokens, nil).Token(context.Background(), TokenRequest{
		ClientID: "partner", ClientSecret: "test-secret-123456",
	})
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	claims, err := tokens.Parse(response.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.PrincipalType != "integration" || claims.IntegrationClientID != 7 || claims.UserID != 0 {
		t.Fatalf("claims = %+v", claims)
	}
	if response.ExpiresIn != 60 || len(response.Permissions) != 2 {
		t.Fatalf("response = %+v", response)
	}
}

func TestTokenRejectsWrongSecret(t *testing.T) {
	hash, _ := password.Hash("test-secret-123456")
	repository := &repositoryStub{client: Client{ID: 7, SecretHash: hash}}
	_, err := NewService(repository, token.New("jwt-secret", time.Minute, time.Hour), nil).Token(
		context.Background(), TokenRequest{ClientID: "partner", ClientSecret: "wrong-secret"},
	)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Token() error = %v", err)
	}
}

func TestCreateWebhookRejectsUnknownEvent(t *testing.T) {
	service := NewService(&repositoryStub{}, token.New("secret", time.Minute, time.Hour), nil)
	_, err := service.CreateWebhook(context.Background(), 1, CreateWebhookRequest{
		URL: "https://partner.example/webhook", SigningSecret: "1234567890123456", Events: []string{"unknown"},
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("CreateWebhook() error = %v", err)
	}
}

func TestImportManualResultsApprovesAndPublishesProtocol(t *testing.T) {
	repository := &repositoryStub{}
	results := &resultServiceStub{table: result.TableResponse{ScoringMode: result.ScoringManual, Status: result.SheetDraft}}
	service := NewService(repository, token.New("secret", time.Minute, time.Hour), results)
	response, err := service.ImportResults(context.Background(), Identity{ClientID: 3, OwnerUserID: 9}, 10, 20,
		ImportResultsRequest{Revision: 1, Entries: []ImportEntry{{EntryID: 30, ResultStatus: result.ResultFinished, Place: int32Pointer(1)}}})
	if err != nil {
		t.Fatalf("ImportResults() error = %v", err)
	}
	if !results.manualSaved || !results.calculated || results.approvedByID != 9 || response.ProtocolURL != "/protocol" {
		t.Fatalf("unexpected import state: %+v / %+v", results, response)
	}
	if len(repository.published) != 1 || repository.published[0].Type != EventProtocolPublished {
		t.Fatalf("published = %+v", repository.published)
	}
}

func int32Pointer(value int32) *int32 { return &value }
