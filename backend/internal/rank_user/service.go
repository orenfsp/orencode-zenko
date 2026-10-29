package rankuser

import (
	"context"
	"strings"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/result"
)

type Service struct {
	repo    Repository
	history ResultHistoryRepository
}

type ResultHistoryRepository interface {
	ListHistory(ctx context.Context, userID int64, disciplineID *int64) ([]result.HistoryItem, error)
}

func NewService(repo Repository, histories ...ResultHistoryRepository) *Service {
	service := &Service{repo: repo}
	if len(histories) > 0 {
		service.history = histories[0]
	}
	return service
}

func (s *Service) Create(ctx context.Context, actorUserID int64, req CreateRequest) (Response, error) {
	athleteUserID := req.AthleteUserID
	if athleteUserID == 0 {
		athleteUserID = actorUserID
	}
	applicationType := strings.ToLower(strings.TrimSpace(req.ApplicationType))
	if actorUserID <= 0 || athleteUserID <= 0 || req.RankID <= 0 || req.DisciplineID <= 0 ||
		req.CompetitionID <= 0 || req.ResultID <= 0 ||
		(applicationType != "assignment" && applicationType != "confirmation") {
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
	validation, err := s.repo.ValidateBasis(ctx, athleteUserID, req.DisciplineID, req.CompetitionID, req.ResultID)
	if err != nil {
		return Response{}, err
	}
	if !validation.Valid() {
		return Response{}, ErrInvalidInput
	}

	application, err := s.repo.Create(
		ctx, athleteUserID, req.RankID, req.DisciplineID, actorUserID,
		applicationType, req.CompetitionID, req.ResultID,
	)
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
	var assignedAt, expiresAt *time.Time
	var err error
	if req.AssignedAt != nil {
		assignedAt, err = parseDate(*req.AssignedAt)
		if err != nil {
			return Response{}, ErrInvalidInput
		}
	}
	if req.ExpiresAt != nil {
		expiresAt, err = parseDate(*req.ExpiresAt)
		if err != nil {
			return Response{}, ErrInvalidInput
		}
	}
	orderNumber := normalizedPointer(req.OrderNumber)
	if status == "approved" {
		if assignedAt == nil || (expiresAt != nil && expiresAt.Before(*assignedAt)) {
			return Response{}, ErrInvalidInput
		}
		rejectionReason = nil
		validEVSK, err := s.repo.ValidateEVSK(ctx, applicationID)
		if err != nil {
			return Response{}, err
		}
		if !validEVSK {
			return Response{}, ErrEVSKNotMet
		}
	} else {
		assignedAt, expiresAt, orderNumber = nil, nil, nil
	}

	application, err := s.repo.Review(ctx, applicationID, reviewerUserID, status, rejectionReason, assignedAt, expiresAt, orderNumber)
	if err != nil {
		return Response{}, err
	}
	return toResponse(application), nil
}

func (s *Service) ListPending(ctx context.Context, limit, offset int32) ([]Response, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrInvalidInput
	}
	applications, err := s.repo.ListPending(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return toResponses(applications), nil
}

func (s *Service) ListHistory(ctx context.Context, athleteUserID int64) ([]HistoryResponse, error) {
	if athleteUserID <= 0 {
		return nil, ErrInvalidInput
	}
	items, err := s.repo.ListHistory(ctx, athleteUserID)
	if err != nil {
		return nil, err
	}
	result := make([]HistoryResponse, 0, len(items))
	for _, item := range items {
		result = append(result, HistoryResponse{
			ID: item.ID, RankID: item.RankID, RankName: item.RankName,
			DisciplineID: item.DisciplineID, DisciplineName: item.DisciplineName,
			ResultID: item.ResultID, AwardedAt: item.AwardedAt, ExpiresAt: item.ExpiresAt,
			OrderNumber: item.OrderNumber, CompetitionID: item.CompetitionID, CompetitionName: item.CompetitionName,
		})
	}
	return result, nil
}

func (s *Service) ListBases(ctx context.Context, actorUserID, athleteUserID, disciplineID int64) ([]BasisResponse, error) {
	if athleteUserID == 0 {
		athleteUserID = actorUserID
	}
	if actorUserID <= 0 || athleteUserID <= 0 || disciplineID <= 0 {
		return nil, ErrInvalidInput
	}
	if actorUserID != athleteUserID {
		allowed, err := s.repo.IsConfirmedCoach(ctx, athleteUserID, actorUserID, disciplineID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrForbidden
		}
	}
	if s.history == nil {
		return nil, ErrInvalidInput
	}
	items, err := s.history.ListHistory(ctx, athleteUserID, &disciplineID)
	if err != nil {
		return nil, err
	}
	bases := make([]BasisResponse, 0, len(items))
	for _, item := range items {
		if item.ResultStatus == nil || *item.ResultStatus != result.ResultFinished {
			continue
		}
		bases = append(bases, BasisResponse{
			ResultID: item.ResultID, CompetitionID: item.CompetitionID,
			CompetitionName: item.CompetitionName, EventID: item.EventID,
			DisciplineID: item.DisciplineID, DisciplineName: item.DisciplineName,
			StartAt: item.StartAt, Place: item.Place, Score: item.Score,
		})
	}
	return bases, nil
}

func (s *Service) ListMine(ctx context.Context, athleteUserID int64) ([]Response, error) {
	if athleteUserID <= 0 {
		return nil, ErrInvalidInput
	}
	applications, err := s.repo.ListByAthlete(ctx, athleteUserID)
	if err != nil {
		return nil, err
	}
	return toResponses(applications), nil
}

func toResponses(applications []Application) []Response {
	result := make([]Response, 0, len(applications))
	for _, application := range applications {
		result = append(result, toResponse(application))
	}
	return result
}

func toResponse(application Application) Response {
	return Response{
		ID:                application.ID,
		AthleteUserID:     application.AthleteUserID,
		RankID:            application.RankID,
		DisciplineID:      application.DisciplineID,
		ApplicationType:   application.ApplicationType,
		CompetitionID:     application.CompetitionID,
		ResultID:          application.ResultID,
		SubmittedByUserID: application.SubmittedByUserID,
		Status:            application.Status,
		ReviewedByUserID:  application.ReviewedByUserID,
		ReviewedAt:        application.ReviewedAt,
		RejectionReason:   application.RejectionReason,
		AssignedAt:        application.AssignedAt,
		ExpiresAt:         application.ExpiresAt,
		OrderNumber:       application.OrderNumber,
		CreatedAt:         application.CreatedAt,
	}
}

func parseDate(value string) (*time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizedPointer(value *string) *string {
	if value == nil {
		return nil
	}
	result := strings.TrimSpace(*value)
	if result == "" {
		return nil
	}
	return &result
}
