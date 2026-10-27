package antidoping

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID, fileID int64) (Certificate, error) {
	if userID <= 0 || fileID <= 0 {
		return Certificate{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, userID, fileID)
}

func (s *Service) ListByUser(ctx context.Context, userID int64) ([]Certificate, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Review(
	ctx context.Context,
	reviewerUserID, certificateID int64,
	request ReviewRequest,
) (Certificate, error) {
	if reviewerUserID <= 0 || certificateID <= 0 {
		return Certificate{}, ErrInvalidInput
	}

	input := ReviewInput{Status: request.Status}
	switch request.Status {
	case StatusValid:
		if request.CompletedAt == nil || request.RejectionReason != nil {
			return Certificate{}, ErrInvalidInput
		}
		completedAt, err := time.Parse(time.DateOnly, *request.CompletedAt)
		if err != nil {
			return Certificate{}, ErrInvalidInput
		}
		input.CompletedAt = completedAt
	case StatusRejected:
		if request.RejectionReason == nil || request.CompletedAt != nil {
			return Certificate{}, ErrInvalidInput
		}
		reason := strings.TrimSpace(*request.RejectionReason)
		if reason == "" || utf8.RuneCountInString(reason) > 500 {
			return Certificate{}, ErrInvalidInput
		}
		input.RejectionReason = reason
	default:
		return Certificate{}, ErrInvalidInput
	}
	return s.repo.Review(ctx, reviewerUserID, certificateID, input)
}

func (s *Service) Expire(ctx context.Context) (int64, error) {
	return s.repo.Expire(ctx)
}
