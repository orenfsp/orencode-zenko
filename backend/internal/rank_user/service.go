package rankuser

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, actorUserID int64, req CreateRequest) (Response, error) {
	athleteUserID := req.AthleteUserID
	if athleteUserID == 0 {
		athleteUserID = actorUserID
	}
	if actorUserID <= 0 || athleteUserID <= 0 || req.RankID <= 0 || req.DisciplineID <= 0 {
		return Response{}, ErrInvalidInput
	}

	if actorUserID == athleteUserID {
		hasCoach, err := s.repo.HasConfirmedCoach(ctx, athleteUserID, req.DisciplineID)
		if err != nil {
			return Response{}, err
		}
		if hasCoach {
			return Response{}, ErrForbidden
		}
	} else {
		isCoach, err := s.repo.IsConfirmedCoach(ctx, athleteUserID, actorUserID, req.DisciplineID)
		if err != nil {
			return Response{}, err
		}
		if !isCoach {
			return Response{}, ErrForbidden
		}
	}

	application, err := s.repo.Create(ctx, athleteUserID, req.RankID, req.DisciplineID, actorUserID)
	if err != nil {
		return Response{}, err
	}
	return toResponse(application), nil
}

func (s *Service) Review(ctx context.Context, reviewerUserID, applicationID int64, req ReviewRequest) (Response, error) {
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if reviewerUserID <= 0 || applicationID <= 0 || (status != "approved" && status != "rejected") {
		return Response{}, ErrInvalidInput
	}
	var rejectionReason *string
	if value := strings.TrimSpace(req.RejectionReason); value != "" {
		if len(value) > 500 {
			return Response{}, ErrInvalidInput
		}
		rejectionReason = &value
	}
	if status == "rejected" && rejectionReason == nil {
		return Response{}, ErrInvalidInput
	}

	application, err := s.repo.Review(ctx, applicationID, reviewerUserID, status, rejectionReason)
	if err != nil {
		return Response{}, err
	}
	return toResponse(application), nil
}

func (s *Service) ListMine(ctx context.Context, athleteUserID int64) ([]Response, error) {
	if athleteUserID <= 0 {
		return nil, ErrInvalidInput
	}
	applications, err := s.repo.ListByAthlete(ctx, athleteUserID)
	if err != nil {
		return nil, err
	}
	result := make([]Response, 0, len(applications))
	for _, application := range applications {
		result = append(result, toResponse(application))
	}
	return result, nil
}

func toResponse(application Application) Response {
	return Response{
		ID:                application.ID,
		AthleteUserID:     application.AthleteUserID,
		RankID:            application.RankID,
		DisciplineID:      application.DisciplineID,
		SubmittedByUserID: application.SubmittedByUserID,
		Status:            application.Status,
		ReviewedByUserID:  application.ReviewedByUserID,
		ReviewedAt:        application.ReviewedAt,
		RejectionReason:   application.RejectionReason,
		CreatedAt:         application.CreatedAt,
	}
}
