package rankuser

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryStub struct {
	hasCoach bool
	isCoach  bool
	created  Application
}

func (r *repositoryStub) HasConfirmedCoach(context.Context, int64, int64) (bool, error) {
	return r.hasCoach, nil
}

func (r *repositoryStub) IsConfirmedCoach(context.Context, int64, int64, int64) (bool, error) {
	return r.isCoach, nil
}

func (r *repositoryStub) ValidateBasis(context.Context, int64, int64, int64, int64) (BasisValidation, error) {
	return BasisValidation{ResultExists: true, AthleteMatches: true, DisciplineMatches: true, CompetitionMatches: true, ResultFinished: true}, nil
}

func (r *repositoryStub) Create(_ context.Context, athleteUserID, rankID, disciplineID, submittedByUserID int64, applicationType string, competitionID, resultID int64) (Application, error) {
	r.created = Application{
		ID:                1,
		AthleteUserID:     athleteUserID,
		RankID:            rankID,
		DisciplineID:      disciplineID,
		ApplicationType:   applicationType,
		CompetitionID:     &competitionID,
		ResultID:          &resultID,
		SubmittedByUserID: submittedByUserID,
		Status:            "pending",
	}
	return r.created, nil
}

func (r *repositoryStub) Review(context.Context, int64, int64, string, *string, *time.Time, *time.Time, *string) (Application, error) {
	return Application{}, nil
}

func (r *repositoryStub) ListPending(context.Context, int32, int32) ([]Application, error) {
	return nil, nil
}
func (r *repositoryStub) ListHistory(context.Context, int64) ([]RankHistory, error) { return nil, nil }

func (r *repositoryStub) ListByAthlete(context.Context, int64) ([]Application, error) {
	return nil, nil
}

func TestCreateAllowsAthleteWithoutConfirmedCoach(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)

	response, err := service.Create(context.Background(), 10, validCreateRequest())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if response.AthleteUserID != 10 || response.SubmittedByUserID != 10 {
		t.Fatalf("response = %#v", response)
	}
}

func TestCreateAllowsAthleteWithConfirmedCoach(t *testing.T) {
	service := NewService(&repositoryStub{hasCoach: true})

	_, err := service.Create(context.Background(), 10, validCreateRequest())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestCreateAllowsConfirmedCoach(t *testing.T) {
	repo := &repositoryStub{isCoach: true}
	service := NewService(repo)

	response, err := service.Create(context.Background(), 20, CreateRequest{
		AthleteUserID:   10,
		RankID:          2,
		DisciplineID:    3,
		ApplicationType: "confirmation",
		CompetitionID:   4,
		ResultID:        5,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if response.AthleteUserID != 10 || response.SubmittedByUserID != 20 {
		t.Fatalf("response = %#v", response)
	}
}

func validCreateRequest() CreateRequest {
	return CreateRequest{RankID: 2, DisciplineID: 3, ApplicationType: "assignment", CompetitionID: 4, ResultID: 5}
}

func TestCreateRejectsMismatchedResult(t *testing.T) {
	service := NewService(&repositoryStub{})
	_, err := service.Create(context.Background(), 10, CreateRequest{RankID: 2, DisciplineID: 3})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error=%v", err)
	}
}
