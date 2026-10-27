package judgecategory

import (
	"context"
	"strings"
	"time"
)

type Service struct {
	repository Repository
	roles      RoleChecker
}

func NewService(repository Repository, roles RoleChecker) *Service {
	return &Service{repository: repository, roles: roles}
}

func (s *Service) Create(ctx context.Context, userID int64, request CreateRequest) (Application, error) {
	if userID <= 0 || request.CategoryID <= 0 || (request.ApplicationType != "assignment" && request.ApplicationType != "confirmation") {
		return Application{}, ErrInvalidInput
	}
	isJudge, err := s.roles.HasRole(ctx, userID, RoleJudge)
	if err != nil {
		return Application{}, err
	}
	if !isJudge {
		return Application{}, ErrForbidden
	}
	exists, err := s.repository.CategoryExists(ctx, request.CategoryID)
	if err != nil {
		return Application{}, err
	}
	if !exists {
		return Application{}, ErrCategoryMissing
	}
	return s.repository.Create(ctx, userID, request)
}

func (s *Service) ListMine(ctx context.Context, userID int64) ([]Application, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	items, err := s.repository.ListMine(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]Application, 0)
	}
	return items, nil
}

func (s *Service) ListPending(ctx context.Context, limit, offset int32) ([]Application, error) {
	if limit < 1 || limit > MaxLimit || offset < 0 {
		return nil, ErrInvalidInput
	}
	items, err := s.repository.ListPending(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]Application, 0)
	}
	return items, nil
}

func (s *Service) Review(ctx context.Context, reviewerID, applicationID int64, request ReviewRequest) (Application, error) {
	if reviewerID <= 0 || applicationID <= 0 || (request.Status != "approved" && request.Status != "rejected") {
		return Application{}, ErrInvalidInput
	}
	request.OrderNumber = normalized(request.OrderNumber)
	request.RejectionReason = normalized(request.RejectionReason)
	input := ReviewInput{Status: request.Status, OrderNumber: request.OrderNumber, RejectionReason: request.RejectionReason}
	var err error
	if request.AssignedAt != nil {
		input.AssignedAt, err = parseDate(*request.AssignedAt)
		if err != nil {
			return Application{}, ErrInvalidInput
		}
	}
	if request.ExpiresAt != nil {
		input.ExpiresAt, err = parseDate(*request.ExpiresAt)
		if err != nil {
			return Application{}, ErrInvalidInput
		}
	}
	if request.Status == "approved" {
		if input.AssignedAt == nil || (input.ExpiresAt != nil && input.ExpiresAt.Before(*input.AssignedAt)) {
			return Application{}, ErrInvalidInput
		}
		input.RejectionReason = nil
	} else {
		if input.RejectionReason == nil {
			return Application{}, ErrInvalidInput
		}
		input.AssignedAt, input.ExpiresAt, input.OrderNumber = nil, nil, nil
	}
	return s.repository.Review(ctx, reviewerID, applicationID, input)
}

func (s *Service) GetCurrent(ctx context.Context, userID int64) (*CurrentCategory, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repository.GetCurrent(ctx, userID)
}

func (s *Service) ListAssignments(ctx context.Context, userID int64) ([]Assignment, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	items, err := s.repository.ListAssignments(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]Assignment, 0)
	}
	return items, nil
}

func parseDate(value string) (*time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalized(value *string) *string {
	if value == nil {
		return nil
	}
	result := strings.TrimSpace(*value)
	if result == "" {
		return nil
	}
	return &result
}
