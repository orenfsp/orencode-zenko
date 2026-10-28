package competitionevent

import (
	"context"
	"errors"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/competition"
)

type repositoryStub struct {
	listPublicFn  func(context.Context, int64) ([]Event, error)
	getPublicFn   func(context.Context, int64, int64) (Event, error)
	listAdminFn   func(context.Context, int64) ([]Event, error)
	getAdminFn    func(context.Context, int64, int64) (Event, error)
	createDraftFn func(context.Context, int64, string, *int64) (Event, error)
	updateDraftFn func(context.Context, int64, int64, string, *int64) (Event, error)
	deleteDraftFn func(context.Context, int64, int64) error
}

func (s *repositoryStub) ListPublic(ctx context.Context, competitionID int64) ([]Event, error) {
	return s.listPublicFn(ctx, competitionID)
}
func (s *repositoryStub) GetPublic(ctx context.Context, competitionID, eventID int64) (Event, error) {
	return s.getPublicFn(ctx, competitionID, eventID)
}
func (s *repositoryStub) ListAdmin(ctx context.Context, competitionID int64) ([]Event, error) {
	return s.listAdminFn(ctx, competitionID)
}
func (s *repositoryStub) GetAdmin(ctx context.Context, competitionID, eventID int64) (Event, error) {
	return s.getAdminFn(ctx, competitionID, eventID)
}
func (s *repositoryStub) CreateDraft(ctx context.Context, competitionID int64, name string, ageGroupID *int64) (Event, error) {
	return s.createDraftFn(ctx, competitionID, name, ageGroupID)
}
func (s *repositoryStub) UpdateDraft(ctx context.Context, competitionID, eventID int64, name string, ageGroupID *int64) (Event, error) {
	return s.updateDraftFn(ctx, competitionID, eventID, name, ageGroupID)
}
func (s *repositoryStub) DeleteDraft(ctx context.Context, competitionID, eventID int64) error {
	return s.deleteDraftFn(ctx, competitionID, eventID)
}

func testEvent() Event {
	ageGroupID, ageGroupName, protocol := int64(3), "Юниоры", "https://example.org/protocol.pdf"
	return Event{ID: 9, CompetitionID: 7, Name: "Личный зачёт", DisciplineID: 2, DisciplineName: "Алгоритмическое программирование",
		AgeGroupID: &ageGroupID, AgeGroupName: &ageGroupName, ProtocolURL: &protocol}
}

func TestServicePublicAndAdminReads(t *testing.T) {
	want := testEvent()
	repo := &repositoryStub{
		listPublicFn: func(context.Context, int64) ([]Event, error) { return []Event{want}, nil },
		getPublicFn:  func(context.Context, int64, int64) (Event, error) { return want, nil },
		listAdminFn:  func(context.Context, int64) ([]Event, error) { return []Event{want}, nil },
		getAdminFn:   func(context.Context, int64, int64) (Event, error) { return want, nil },
	}
	service := NewService(repo)

	public, err := service.ListPublic(context.Background(), 7)
	if err != nil || len(public) != 1 || public[0].ProtocolURL == nil || *public[0].ProtocolURL == "" {
		t.Fatalf("ListPublic() = %#v, %v", public, err)
	}
	if _, err := service.GetPublic(context.Background(), 7, 9); err != nil {
		t.Fatalf("GetPublic() error = %v", err)
	}
	admin, err := service.ListAdmin(context.Background(), 7)
	if err != nil || len(admin) != 1 {
		t.Fatalf("ListAdmin() = %#v, %v", admin, err)
	}
	if _, err := service.GetAdmin(context.Background(), 7, 9); err != nil {
		t.Fatalf("GetAdmin() error = %v", err)
	}
}

func TestServicePropagatesReadVisibilityAndOwnershipErrors(t *testing.T) {
	repo := &repositoryStub{
		listPublicFn: func(context.Context, int64) ([]Event, error) { return nil, ErrCompetitionMissing },
		getPublicFn:  func(context.Context, int64, int64) (Event, error) { return Event{}, ErrEventMissing },
	}
	service := NewService(repo)
	if _, err := service.ListPublic(context.Background(), 7); !errors.Is(err, ErrCompetitionMissing) {
		t.Fatalf("ListPublic() error = %v", err)
	}
	if _, err := service.GetPublic(context.Background(), 7, 99); !errors.Is(err, ErrEventMissing) {
		t.Fatalf("GetPublic() error = %v", err)
	}
}

func TestServiceCreateDerivesDisciplineInRepository(t *testing.T) {
	ageGroupID := int64(3)
	repo := &repositoryStub{createDraftFn: func(_ context.Context, competitionID int64, name string, got *int64) (Event, error) {
		if competitionID != 7 || name != "Личный зачёт" || got == nil || *got != ageGroupID {
			t.Fatalf("CreateDraft(%d, %v)", competitionID, got)
		}
		return testEvent(), nil
	}}
	created, err := NewService(repo).Create(context.Background(), 7, CreateRequest{Name: "Личный зачёт", AgeGroupID: &ageGroupID})
	if err != nil || created.Discipline.ID != 2 {
		t.Fatalf("Create() = %#v, %v", created, err)
	}
}

func TestServiceRejectsInvalidAgeGroupID(t *testing.T) {
	invalid := int64(0)
	service := NewService(&repositoryStub{})
	if _, err := service.Create(context.Background(), 7, CreateRequest{AgeGroupID: &invalid}); !errors.Is(err, ErrInvalidCreate) {
		t.Fatalf("Create() error = %v", err)
	}
	request := UpdateRequest{AgeGroupID: competition.PatchField[int64]{Present: true, Value: -1}}
	if _, err := service.Update(context.Background(), 7, 9, request); !errors.Is(err, ErrInvalidUpdate) {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestServiceUpdateAndClearAgeGroup(t *testing.T) {
	tests := []struct {
		name    string
		field   competition.PatchField[int64]
		wantNil bool
	}{
		{name: "replace", field: competition.PatchField[int64]{Present: true, Value: 4}},
		{name: "clear with null", field: competition.PatchField[int64]{Present: true, Null: true}, wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{getAdminFn: func(context.Context, int64, int64) (Event, error) { return testEvent(), nil }, updateDraftFn: func(_ context.Context, competitionID, eventID int64, name string, ageGroupID *int64) (Event, error) {
				if competitionID != 7 || eventID != 9 || (ageGroupID == nil) != tt.wantNil {
					t.Fatalf("UpdateDraft(%d, %d, %v)", competitionID, eventID, ageGroupID)
				}
				return testEvent(), nil
			}}
			if _, err := NewService(repo).Update(context.Background(), 7, 9, UpdateRequest{AgeGroupID: tt.field}); err != nil {
				t.Fatalf("Update() error = %v", err)
			}
		})
	}
}

func TestServiceMutationConflicts(t *testing.T) {
	for _, want := range []error{ErrNotDraft, ErrAgeGroupMissing, ErrHasApplications} {
		t.Run(want.Error(), func(t *testing.T) {
			repo := &repositoryStub{
				getAdminFn:    func(context.Context, int64, int64) (Event, error) { return testEvent(), nil },
				updateDraftFn: func(context.Context, int64, int64, string, *int64) (Event, error) { return Event{}, want },
				deleteDraftFn: func(context.Context, int64, int64) error { return want },
			}
			request := UpdateRequest{AgeGroupID: competition.PatchField[int64]{Present: true, Null: true}}
			if _, err := NewService(repo).Update(context.Background(), 7, 9, request); !errors.Is(err, want) {
				t.Fatalf("Update() error = %v, want %v", err, want)
			}
			if err := NewService(repo).Delete(context.Background(), 7, 9); !errors.Is(err, want) {
				t.Fatalf("Delete() error = %v, want %v", err, want)
			}
		})
	}
}

func TestServiceDeleteSuccess(t *testing.T) {
	repo := &repositoryStub{deleteDraftFn: func(context.Context, int64, int64) error { return nil }}
	if err := NewService(repo).Delete(context.Background(), 7, 9); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestServiceRejectsEmptyPatch(t *testing.T) {
	if _, err := NewService(&repositoryStub{}).Update(context.Background(), 7, 9, UpdateRequest{}); !errors.Is(err, ErrInvalidUpdate) {
		t.Fatalf("Update() error = %v", err)
	}
}
