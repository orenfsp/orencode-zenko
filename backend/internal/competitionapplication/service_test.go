package competitionapplication

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryStub struct {
	event       EventContext
	application Application
	items       []Application
	confirmed   bool
	createErr   error
	reviewed    string
}

func (r *repositoryStub) GetEventContext(context.Context, int64) (EventContext, error) { return r.event, nil }
func (r *repositoryStub) IsConfirmedCoach(context.Context, int64, int64, int64) (bool, error) { return r.confirmed, nil }
func (r *repositoryStub) Create(_ context.Context, eventID, athleteID, submitterID int64) (Application, error) {
	if r.createErr != nil { return Application{}, r.createErr }
	result := r.application; result.EventID, result.AthleteUserID, result.SubmittedByUserID = eventID, athleteID, submitterID
	return result, nil
}
func (r *repositoryStub) Get(context.Context, int64) (Application, error) { return r.application, nil }
func (r *repositoryStub) ListAccessible(context.Context, int64) ([]Application, error) { return r.items, nil }
func (r *repositoryStub) ListForReview(context.Context, ListFilter) ([]Application, error) { return r.items, nil }
func (r *repositoryStub) Review(_ context.Context, _ int64, _ int64, status string) (Application, error) {
	r.reviewed = status; result := r.application; result.Status = status; return result, nil
}
func (r *repositoryStub) Withdraw(context.Context, int64, int64) (Application, error) {
	result := r.application; result.Status = "withdrawn"; return result, nil
}

func openRepository(now time.Time) *repositoryStub {
	return &repositoryStub{
		event: EventContext{EventID: 10, DisciplineID: 20, CompetitionStatus: "registration_open", RegistrationStartsAt: now.Add(-time.Hour), RegistrationEndsAt: now.Add(time.Hour)},
		application: Application{ID: 30, Status: "pending", RegistrationEndsAt: now.Add(time.Hour)},
	}
}

func newTestService(repository Repository, now time.Time) *Service {
	service := NewService(repository); service.now = func() time.Time { return now }; return service
}

func TestCreateForSelf(t *testing.T) {
	now := time.Now(); repository := openRepository(now); service := newTestService(repository, now)
	result, err := service.Create(context.Background(), 7, CreateRequest{EventID: 10})
	if err != nil { t.Fatal(err) }
	if result.AthleteUserID != 7 || result.SubmittedByUserID != 7 || !result.CanWithdraw { t.Fatalf("unexpected application: %+v", result) }
}

func TestCreateForConfirmedAthlete(t *testing.T) {
	now := time.Now(); repository := openRepository(now); repository.confirmed = true; service := newTestService(repository, now)
	result, err := service.Create(context.Background(), 8, CreateRequest{EventID: 10, AthleteUserID: 7})
	if err != nil { t.Fatal(err) }
	if result.AthleteUserID != 7 || result.SubmittedByUserID != 8 { t.Fatalf("unexpected application: %+v", result) }
}

func TestCreateForUnrelatedAthleteIsForbidden(t *testing.T) {
	now := time.Now(); service := newTestService(openRepository(now), now)
	_, err := service.Create(context.Background(), 8, CreateRequest{EventID: 10, AthleteUserID: 7})
	if !errors.Is(err, ErrForbidden) { t.Fatalf("expected forbidden, got %v", err) }
}

func TestCreateAfterRegistrationIsClosed(t *testing.T) {
	now := time.Now(); repository := openRepository(now); repository.event.RegistrationEndsAt = now.Add(-time.Minute)
	_, err := newTestService(repository, now).Create(context.Background(), 7, CreateRequest{EventID: 10})
	if !errors.Is(err, ErrRegistrationClosed) { t.Fatalf("expected closed registration, got %v", err) }
}

func TestCreateReturnsDuplicateError(t *testing.T) {
	now := time.Now(); repository := openRepository(now); repository.createErr = ErrConflict
	_, err := newTestService(repository, now).Create(context.Background(), 7, CreateRequest{EventID: 10})
	if !errors.Is(err, ErrConflict) { t.Fatalf("expected conflict, got %v", err) }
}

func TestReviewAcceptsOnlyPendingDecisionStatuses(t *testing.T) {
	now := time.Now(); repository := openRepository(now); service := newTestService(repository, now)
	if _, err := service.Review(context.Background(), 1, 30, "approved"); err != nil { t.Fatal(err) }
	if repository.reviewed != "approved" { t.Fatalf("unexpected status %q", repository.reviewed) }
	if _, err := service.Review(context.Background(), 1, 30, "withdrawn"); !errors.Is(err, ErrInvalidInput) { t.Fatalf("expected invalid input, got %v", err) }
}

func TestWithdrawOnlyOwnPendingApplication(t *testing.T) {
	now := time.Now(); repository := openRepository(now); repository.application.AthleteUserID = 7
	result, err := newTestService(repository, now).Withdraw(context.Background(), 7, 30)
	if err != nil { t.Fatal(err) }
	if result.Status != "withdrawn" { t.Fatalf("unexpected status %q", result.Status) }
}
