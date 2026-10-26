package competition

import (
	"context"
	"errors"
	"testing"
	"time"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

type repositoryStub struct {
	items      []Competition
	item       Competition
	listErr    error
	getErr     error
	seenLimit  int32
	seenOffset int32
	seenID     int64
}

func (s *repositoryStub) ListPublished(_ context.Context, limit, offset int32) ([]Competition, error) {
	s.seenLimit = limit
	s.seenOffset = offset
	return s.items, s.listErr
}

func (s *repositoryStub) GetPublished(_ context.Context, id int64) (Competition, error) {
	s.seenID = id
	return s.item, s.getErr
}

func testCompetition() Competition {
	return Competition{
		ID: 7, Name: "Кубок Оренбуржья", DisciplineID: 2,
		DisciplineName:    "Алгоритмическое программирование",
		CompetitionTypeID: 3, CompetitionTypeName: "Кубок",
		CompetitionLevelID: 4, Format: "mixed", Status: "published",
		StartAt:              time.Date(2026, time.November, 10, 8, 0, 0, 0, time.UTC),
		EndAt:                time.Date(2026, time.November, 11, 16, 0, 0, 0, time.UTC),
		RegistrationStartsAt: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		RegistrationEndsAt:   time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestServiceListPublishedCompetitions(t *testing.T) {
	repo := &repositoryStub{items: []Competition{testCompetition()}}
	service := NewService(repo)

	result, err := service.List(context.Background(), 25, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.seenLimit != 25 || repo.seenOffset != 10 {
		t.Fatalf("repository pagination = (%d, %d), want (25, 10)", repo.seenLimit, repo.seenOffset)
	}
	if len(result) != 1 || result[0].ID != 7 || result[0].CompetitionType.Name != "Кубок" {
		t.Fatalf("List() result = %#v", result)
	}
}

func TestServiceListReturnsEmptyJSONReadySlice(t *testing.T) {
	result, err := NewService(&repositoryStub{}).List(context.Background(), DefaultLimit, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result == nil || len(result) != 0 {
		t.Fatalf("List() result = %#v, want non-nil empty slice", result)
	}
}

func TestServiceListRejectsInvalidPagination(t *testing.T) {
	tests := []struct {
		name   string
		limit  int32
		offset int32
	}{
		{name: "zero limit", limit: 0},
		{name: "over maximum", limit: MaxLimit + 1},
		{name: "negative offset", limit: DefaultLimit, offset: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(&repositoryStub{}).List(context.Background(), tt.limit, tt.offset)
			if !errors.Is(err, competitionerrs.ErrInvalidPagination) {
				t.Fatalf("List() error = %v, want ErrInvalidPagination", err)
			}
		})
	}
}

func TestServiceListPropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	_, err := NewService(&repositoryStub{listErr: wantErr}).List(context.Background(), DefaultLimit, 0)
	if !errors.Is(err, wantErr) {
		t.Fatalf("List() error = %v, want %v", err, wantErr)
	}
}

func TestServiceGetPublishedCompetition(t *testing.T) {
	repo := &repositoryStub{item: testCompetition()}
	result, err := NewService(repo).Get(context.Background(), 7)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if repo.seenID != 7 || result.ID != 7 || result.Discipline.Name == "" {
		t.Fatalf("Get() result = %#v, seen ID = %d", result, repo.seenID)
	}
}

func TestServiceGetMissingCompetition(t *testing.T) {
	_, err := NewService(&repositoryStub{getErr: competitionerrs.ErrNotFound}).Get(context.Background(), 404)
	if !errors.Is(err, competitionerrs.ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestServiceGetUnpublishedCompetitionIsNotAccessible(t *testing.T) {
	// The repository's published-only SQL reports drafts exactly like missing rows.
	_, err := NewService(&repositoryStub{getErr: competitionerrs.ErrNotFound}).Get(context.Background(), 8)
	if !errors.Is(err, competitionerrs.ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestServiceGetRejectsInvalidID(t *testing.T) {
	_, err := NewService(&repositoryStub{}).Get(context.Background(), 0)
	if !errors.Is(err, competitionerrs.ErrInvalidID) {
		t.Fatalf("Get() error = %v, want ErrInvalidID", err)
	}
}
