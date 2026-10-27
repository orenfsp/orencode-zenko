package competition

import (
	"context"
	"errors"
	"testing"
	"time"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

type repositoryStub struct {
	items          []Competition
	item           Competition
	created        Competition
	listErr        error
	getErr         error
	referenceErr   error
	createErr      error
	updateErr      error
	seenLimit      int32
	seenOffset     int32
	seenID         int64
	seenReferences CompetitionReferences
	seenCreate     CreateCompetition
	seenUpdate     UpdateCompetition
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

func (s *repositoryStub) ValidateReferences(_ context.Context, references CompetitionReferences) error {
	s.seenReferences = references
	return s.referenceErr
}

func (s *repositoryStub) Create(_ context.Context, input CreateCompetition) (Competition, error) {
	s.seenCreate = input
	return s.created, s.createErr
}

func (s *repositoryStub) GetByID(_ context.Context, id int64) (Competition, error) {
	s.seenID = id
	return s.item, s.getErr
}

func (s *repositoryStub) UpdateDraft(_ context.Context, input UpdateCompetition) (Competition, error) {
	s.seenUpdate = input
	return s.created, s.updateErr
}

func testCompetition() Competition {
	venueAddress := "Оренбург, ул. Советская, 1"
	onlineURL := "https://contest.example.org"
	return Competition{
		ID: 7, Name: "Кубок Оренбуржья", DisciplineID: 2,
		DisciplineName:    "Алгоритмическое программирование",
		CompetitionTypeID: 3, CompetitionTypeName: "Кубок",
		CompetitionLevelID: 4, Format: "mixed", Status: "published",
		StartAt:              time.Date(2026, time.November, 10, 8, 0, 0, 0, time.UTC),
		EndAt:                time.Date(2026, time.November, 11, 16, 0, 0, 0, time.UTC),
		RegistrationStartsAt: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		RegistrationEndsAt:   time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC),
		VenueAddress:         &venueAddress,
		OnlineURL:            &onlineURL,
	}
}

func validCreateRequest() CreateRequest {
	venue := " Оренбург, ул. Советская, 1 "
	return CreateRequest{
		Name: " Кубок Оренбуржья ", DisciplineID: 2, CompetitionTypeID: 3,
		CompetitionLevelID: 4, Format: " MIXED ",
		StartAt:              time.Date(2026, time.November, 10, 8, 0, 0, 0, time.UTC),
		EndAt:                time.Date(2026, time.November, 11, 16, 0, 0, 0, time.UTC),
		RegistrationStartsAt: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		RegistrationEndsAt:   time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC),
		VenueAddress:         &venue, OnlineURL: stringPointer("https://contest.example.org"),
	}
}

func stringPointer(value string) *string { return &value }

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

func TestServiceCreatesDraftCompetition(t *testing.T) {
	created := testCompetition()
	created.Status = "draft"
	repo := &repositoryStub{created: created}

	result, err := NewService(repo).Create(context.Background(), 42, validCreateRequest())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.Status != "draft" || result.ID != created.ID {
		t.Fatalf("Create() result = %#v", result)
	}
	if repo.seenCreate.CreatedBy != 42 || repo.seenCreate.Name != "Кубок Оренбуржья" ||
		repo.seenCreate.Format != "mixed" || repo.seenCreate.VenueAddress == nil ||
		*repo.seenCreate.VenueAddress != "Оренбург, ул. Советская, 1" {
		t.Fatalf("Create() repository input = %#v", repo.seenCreate)
	}
	if repo.seenReferences.DisciplineID != 2 || repo.seenReferences.CompetitionLevelID != 4 {
		t.Fatalf("validated references = %#v", repo.seenReferences)
	}
}

func TestServiceCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		change func(*CreateRequest)
	}{
		{name: "missing name", change: func(req *CreateRequest) { req.Name = " " }},
		{name: "unknown format", change: func(req *CreateRequest) { req.Format = "hybrid" }},
		{name: "end before start", change: func(req *CreateRequest) { req.EndAt = req.StartAt }},
		{name: "registration after start", change: func(req *CreateRequest) { req.RegistrationEndsAt = req.StartAt.Add(time.Hour) }},
		{name: "mixed without venue", change: func(req *CreateRequest) { req.VenueAddress = nil }},
		{name: "invalid online URL", change: func(req *CreateRequest) { req.OnlineURL = stringPointer("javascript:alert(1)") }},
		{name: "calendar number without plan", change: func(req *CreateRequest) { req.CalendarPlanNumber = stringPointer("15") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validCreateRequest()
			tt.change(&req)
			repo := &repositoryStub{}
			_, err := NewService(repo).Create(context.Background(), 42, req)
			if !errors.Is(err, competitionerrs.ErrInvalidCreate) {
				t.Fatalf("Create() error = %v, want ErrInvalidCreate", err)
			}
			if repo.seenCreate.Name != "" {
				t.Fatal("repository Create() called for invalid input")
			}
		})
	}
}

func TestServiceCreateRejectsMissingAuthorizedEmployee(t *testing.T) {
	_, err := NewService(&repositoryStub{}).Create(context.Background(), 0, validCreateRequest())
	if !errors.Is(err, competitionerrs.ErrForbidden) {
		t.Fatalf("Create() error = %v, want ErrForbidden", err)
	}
}

func TestServiceCreateRejectsInvalidReferences(t *testing.T) {
	repo := &repositoryStub{referenceErr: competitionerrs.ErrInvalidReference}
	_, err := NewService(repo).Create(context.Background(), 42, validCreateRequest())
	if !errors.Is(err, competitionerrs.ErrInvalidReference) {
		t.Fatalf("Create() error = %v, want ErrInvalidReference", err)
	}
	if repo.seenCreate.Name != "" {
		t.Fatal("repository Create() called after invalid references")
	}
}

func TestServiceCreatePropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	_, err := NewService(&repositoryStub{createErr: wantErr}).Create(context.Background(), 42, validCreateRequest())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want %v", err, wantErr)
	}
}

func TestServiceUpdateChangesOnlyProvidedFields(t *testing.T) {
	current := testCompetition()
	current.Status = "draft"
	current.Description = stringPointer("Старое описание")
	updated := current
	updated.Name = "Новое название"
	repo := &repositoryStub{item: current, created: updated}
	req := UpdateRequest{Name: PatchField[string]{Present: true, Value: " Новое название "}}

	result, err := NewService(repo).Update(context.Background(), current.ID, req)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if repo.seenUpdate.Name != "Новое название" {
		t.Fatalf("updated name = %q", repo.seenUpdate.Name)
	}
	if repo.seenUpdate.Description == nil || *repo.seenUpdate.Description != "Старое описание" {
		t.Fatalf("unmentioned description changed: %#v", repo.seenUpdate.Description)
	}
	if !repo.seenUpdate.StartAt.Equal(current.StartAt) || result.Name != "Новое название" {
		t.Fatal("unmentioned dates changed or response was not mapped")
	}
}

func TestServiceUpdateClearsNullableField(t *testing.T) {
	current := testCompetition()
	current.Status = "draft"
	current.Description = stringPointer("Удалить")
	updated := current
	updated.Description = nil
	repo := &repositoryStub{item: current, created: updated}

	_, err := NewService(repo).Update(context.Background(), current.ID, UpdateRequest{
		Description: PatchField[string]{Present: true, Null: true},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if repo.seenUpdate.Description != nil {
		t.Fatalf("description = %#v, want nil", repo.seenUpdate.Description)
	}
}

func TestServiceUpdateRejectsEmptyPatch(t *testing.T) {
	_, err := NewService(&repositoryStub{}).Update(context.Background(), 7, UpdateRequest{})
	if !errors.Is(err, competitionerrs.ErrInvalidUpdate) {
		t.Fatalf("Update() error = %v, want ErrInvalidUpdate", err)
	}
}

func TestServiceUpdateRejectsNullRequiredField(t *testing.T) {
	current := testCompetition()
	current.Status = "draft"
	_, err := NewService(&repositoryStub{item: current}).Update(context.Background(), 7, UpdateRequest{
		Name: PatchField[string]{Present: true, Null: true},
	})
	if !errors.Is(err, competitionerrs.ErrInvalidUpdate) {
		t.Fatalf("Update() error = %v, want ErrInvalidUpdate", err)
	}
}

func TestServiceUpdateMissingCompetition(t *testing.T) {
	_, err := NewService(&repositoryStub{getErr: competitionerrs.ErrNotFound}).Update(
		context.Background(), 404, UpdateRequest{Name: PatchField[string]{Present: true, Value: "Новое"}},
	)
	if !errors.Is(err, competitionerrs.ErrNotFound) {
		t.Fatalf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestServiceUpdateRejectsPublishedCompetition(t *testing.T) {
	current := testCompetition()
	current.Status = "published"
	_, err := NewService(&repositoryStub{item: current}).Update(
		context.Background(), current.ID, UpdateRequest{Name: PatchField[string]{Present: true, Value: "Новое"}},
	)
	if !errors.Is(err, competitionerrs.ErrNotDraft) {
		t.Fatalf("Update() error = %v, want ErrNotDraft", err)
	}
}

func TestServiceUpdateRejectsInvalidFinalDates(t *testing.T) {
	current := testCompetition()
	current.Status = "draft"
	_, err := NewService(&repositoryStub{item: current}).Update(context.Background(), current.ID, UpdateRequest{
		EndAt: PatchField[time.Time]{Present: true, Value: current.StartAt},
	})
	if !errors.Is(err, competitionerrs.ErrInvalidUpdate) {
		t.Fatalf("Update() error = %v, want ErrInvalidUpdate", err)
	}
}

func TestServiceUpdatePropagatesRepositoryError(t *testing.T) {
	current := testCompetition()
	current.Status = "draft"
	wantErr := errors.New("update failed")
	_, err := NewService(&repositoryStub{item: current, updateErr: wantErr}).Update(
		context.Background(), current.ID, UpdateRequest{Name: PatchField[string]{Present: true, Value: "Новое"}},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want %v", err, wantErr)
	}
}
