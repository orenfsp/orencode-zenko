package rating

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryStub struct {
	disciplines map[int64]bool
	ageGroups   map[int64]bool
	roles       map[int64]map[string]bool
	saved       *SaveSelectionRequest
	actorID     int64
	selection   Selection
	err         error
}

func (r *repositoryStub) List(context.Context, Filter) ([]Entry, error) { return nil, r.err }
func (r *repositoryStub) DisciplineExists(_ context.Context, id int64) (bool, error) {
	return r.disciplines[id], r.err
}
func (r *repositoryStub) AgeGroupExists(_ context.Context, id int64) (bool, error) {
	return r.ageGroups[id], r.err
}
func (r *repositoryStub) UserHasRole(_ context.Context, userID int64, roleCode string) (bool, error) {
	return r.roles[userID][roleCode], r.err
}
func (r *repositoryStub) GetSelection(context.Context, SelectionScope) (Selection, error) {
	return r.selection, r.err
}
func (r *repositoryStub) SaveSelection(_ context.Context, actorID int64, request SaveSelectionRequest) (Selection, error) {
	r.actorID = actorID
	r.saved = &request
	return r.selection, r.err
}
func (r *repositoryStub) LockSelection(context.Context, int64) (Selection, error) {
	return r.selection, r.err
}

func TestSaveSelectionPreservesManualOrder(t *testing.T) {
	captainID, coachID := int64(20), int64(30)
	repository := &repositoryStub{
		disciplines: map[int64]bool{2: true},
		ageGroups:   map[int64]bool{3: true},
		roles: map[int64]map[string]bool{
			10: {"participant": true}, 20: {"participant": true}, 30: {"coach": true},
		},
		selection: Selection{ID: 7},
	}
	request := SaveSelectionRequest{
		SeasonYear: int32(time.Now().Year()), DisciplineID: 2, AgeGroupID: 3,
		MemberIDs: []int64{20, 10}, CaptainUserID: &captainID, CoachUserID: &coachID,
	}

	selection, err := NewService(repository).SaveSelection(context.Background(), 1, request)
	if err != nil {
		t.Fatalf("SaveSelection() error = %v", err)
	}
	if selection.ID != 7 || repository.actorID != 1 || repository.saved == nil {
		t.Fatalf("selection/repository = %+v / %+v", selection, repository)
	}
	if repository.saved.MemberIDs[0] != 20 || repository.saved.MemberIDs[1] != 10 {
		t.Fatalf("member order changed: %v", repository.saved.MemberIDs)
	}
}

func TestSaveSelectionRejectsCaptainOutsideMembers(t *testing.T) {
	captainID := int64(99)
	service := NewService(&repositoryStub{})

	_, err := service.SaveSelection(context.Background(), 1, SaveSelectionRequest{
		SeasonYear: int32(time.Now().Year()), DisciplineID: 2, AgeGroupID: 3,
		MemberIDs: []int64{10}, CaptainUserID: &captainID,
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("SaveSelection() error = %v, want ErrInvalidInput", err)
	}
}

func TestSaveSelectionRejectsDuplicateMembers(t *testing.T) {
	service := NewService(&repositoryStub{})

	_, err := service.SaveSelection(context.Background(), 1, SaveSelectionRequest{
		SeasonYear: int32(time.Now().Year()), DisciplineID: 2, AgeGroupID: 3,
		MemberIDs: []int64{10, 10},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("SaveSelection() error = %v, want ErrInvalidInput", err)
	}
}

func TestSaveSelectionRejectsUserWithoutParticipantRole(t *testing.T) {
	repository := &repositoryStub{
		disciplines: map[int64]bool{2: true}, ageGroups: map[int64]bool{3: true},
		roles: map[int64]map[string]bool{10: {}},
	}
	_, err := NewService(repository).SaveSelection(context.Background(), 1, SaveSelectionRequest{
		SeasonYear: int32(time.Now().Year()), DisciplineID: 2, AgeGroupID: 3, MemberIDs: []int64{10},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("SaveSelection() error = %v, want ErrInvalidInput", err)
	}
}
