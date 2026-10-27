package antidoping

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryStub struct {
	reviewInput ReviewInput
	reviewCalls int
}

func (r *repositoryStub) Create(context.Context, int64, int64) (Certificate, error) {
	return Certificate{}, nil
}

func (r *repositoryStub) ListByUser(context.Context, int64) ([]Certificate, error) {
	return nil, nil
}

func (r *repositoryStub) Review(_ context.Context, _, _ int64, input ReviewInput) (Certificate, error) {
	r.reviewCalls++
	r.reviewInput = input
	return Certificate{Status: input.Status}, nil
}

func (r *repositoryStub) Expire(context.Context) (int64, error) { return 0, nil }

func TestReviewValidParsesCompletedAt(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)
	date := "2026-03-12"

	certificate, err := service.Review(context.Background(), 2, 25, ReviewRequest{
		Status:      StatusValid,
		CompletedAt: &date,
	})
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if certificate.Status != StatusValid || repo.reviewCalls != 1 {
		t.Fatalf("Review() certificate = %#v, calls = %d", certificate, repo.reviewCalls)
	}
	if got := repo.reviewInput.CompletedAt.Format(time.DateOnly); got != date {
		t.Fatalf("completed_at = %q, want %q", got, date)
	}
}

func TestReviewRejectedTrimsReason(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)
	reason := "  Документ нечитаем  "

	_, err := service.Review(context.Background(), 2, 25, ReviewRequest{
		Status:          StatusRejected,
		RejectionReason: &reason,
	})
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if repo.reviewInput.RejectionReason != "Документ нечитаем" {
		t.Fatalf("rejection_reason = %q", repo.reviewInput.RejectionReason)
	}
}

func TestReviewRejectsInvalidPayloads(t *testing.T) {
	validDate := "2026-03-12"
	reason := "Причина"
	empty := "  "
	long := string(make([]rune, 501))
	tests := []ReviewRequest{
		{Status: StatusManualReview},
		{Status: StatusValid},
		{Status: StatusValid, CompletedAt: &reason},
		{Status: StatusValid, CompletedAt: &validDate, RejectionReason: &reason},
		{Status: StatusRejected},
		{Status: StatusRejected, RejectionReason: &empty},
		{Status: StatusRejected, RejectionReason: &long},
		{Status: StatusRejected, CompletedAt: &validDate, RejectionReason: &reason},
	}
	for _, request := range tests {
		repo := &repositoryStub{}
		_, err := NewService(repo).Review(context.Background(), 2, 25, request)
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("Review(%#v) error = %v, want ErrInvalidInput", request, err)
		}
		if repo.reviewCalls != 0 {
			t.Errorf("Review(%#v) called repository", request)
		}
	}
}
