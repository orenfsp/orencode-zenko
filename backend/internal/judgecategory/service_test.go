package judgecategory

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	categoryExists bool
	created        Application
	items          []Application
	assignments    []Assignment
	current        *CurrentCategory
	reviewInput    ReviewInput
	err            error
}

func (r *repositoryStub) CategoryExists(context.Context, int64) (bool, error) {
	return r.categoryExists, r.err
}
func (r *repositoryStub) Create(context.Context, int64, CreateRequest) (Application, error) {
	return r.created, r.err
}
func (r *repositoryStub) ListMine(context.Context, int64) ([]Application, error) {
	return r.items, r.err
}
func (r *repositoryStub) ListPending(context.Context, int32, int32) ([]Application, error) {
	return r.items, r.err
}
func (r *repositoryStub) Review(_ context.Context, _, _ int64, input ReviewInput) (Application, error) {
	r.reviewInput = input
	return r.created, r.err
}
func (r *repositoryStub) GetCurrent(context.Context, int64) (*CurrentCategory, error) {
	return r.current, r.err
}
func (r *repositoryStub) ListAssignments(context.Context, int64) ([]Assignment, error) {
	return r.assignments, r.err
}

type roleStub struct {
	allowed bool
	err     error
}

func (r roleStub) HasRole(context.Context, int64, string) (bool, error) { return r.allowed, r.err }

func TestCreateJudgeCategoryApplication(t *testing.T) {
	repository := &repositoryStub{categoryExists: true, created: Application{ID: 9}}
	application, err := NewService(repository, roleStub{allowed: true}).Create(context.Background(), 4, CreateRequest{
		CategoryID: 2, ApplicationType: "assignment",
	})
	if err != nil || application.ID != 9 {
		t.Fatalf("application=%#v err=%v", application, err)
	}
}

func TestCreateRequiresJudgeRole(t *testing.T) {
	_, err := NewService(&repositoryStub{}, roleStub{}).Create(context.Background(), 4, CreateRequest{CategoryID: 2, ApplicationType: "assignment"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("error=%v", err)
	}
}

func TestReviewApprovedParsesDates(t *testing.T) {
	assigned, expires := "2026-10-01", "2027-10-01"
	repository := &repositoryStub{}
	_, err := NewService(repository, roleStub{}).Review(context.Background(), 1, 2, ReviewRequest{
		Status: "approved", AssignedAt: &assigned, ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repository.reviewInput.AssignedAt == nil || repository.reviewInput.ExpiresAt == nil {
		t.Fatalf("input=%#v", repository.reviewInput)
	}
}

func TestReviewRejectedRequiresReason(t *testing.T) {
	_, err := NewService(&repositoryStub{}, roleStub{}).Review(context.Background(), 1, 2, ReviewRequest{Status: "rejected"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error=%v", err)
	}
}

func TestListMineReturnsEmptyArray(t *testing.T) {
	items, err := NewService(&repositoryStub{}, roleStub{}).ListMine(context.Background(), 4)
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
