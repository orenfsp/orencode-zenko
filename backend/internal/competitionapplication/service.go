package competitionapplication

import (
	"context"
	"strings"
	"time"
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) Create(ctx context.Context, actorUserID int64, request CreateRequest) (Application, error) {
	athleteUserID := request.AthleteUserID
	if athleteUserID == 0 {
		athleteUserID = actorUserID
	}
	if actorUserID <= 0 || athleteUserID <= 0 || request.EventID <= 0 {
		return Application{}, ErrInvalidInput
	}

	event, err := s.repository.GetEventContext(ctx, request.EventID)
	if err != nil {
		return Application{}, err
	}
	now := s.now()
	if event.CompetitionStatus != "registration_open" || now.Before(event.RegistrationStartsAt) || now.After(event.RegistrationEndsAt) {
		return Application{}, ErrRegistrationClosed
	}
	if actorUserID != athleteUserID {
		confirmed, checkErr := s.repository.IsConfirmedCoach(ctx, athleteUserID, actorUserID, event.DisciplineID)
		if checkErr != nil {
			return Application{}, checkErr
		}
		if !confirmed {
			return Application{}, ErrForbidden
		}
	}
	application, err := s.repository.Create(ctx, request.EventID, athleteUserID, actorUserID)
	if err != nil {
		return Application{}, err
	}
	application.CanWithdraw = application.AthleteUserID == actorUserID && canWithdraw(application, now)
	return application, nil
}

func (s *Service) ListAccessible(ctx context.Context, actorUserID int64) ([]Application, error) {
	if actorUserID <= 0 {
		return nil, ErrInvalidInput
	}
	applications, err := s.repository.ListAccessible(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	for index := range applications {
		applications[index].CanWithdraw = applications[index].AthleteUserID == actorUserID && canWithdraw(applications[index], now)
	}
	return applications, nil
}

func (s *Service) GetAccessible(ctx context.Context, actorUserID, applicationID int64) (Application, error) {
	if actorUserID <= 0 || applicationID <= 0 {
		return Application{}, ErrInvalidInput
	}
	application, err := s.repository.Get(ctx, applicationID)
	if err != nil {
		return Application{}, err
	}
	if application.AthleteUserID != actorUserID {
		confirmed, checkErr := s.repository.IsConfirmedCoach(ctx, application.AthleteUserID, actorUserID, application.DisciplineID)
		if checkErr != nil {
			return Application{}, checkErr
		}
		if !confirmed {
			return Application{}, ErrForbidden
		}
	}
	application.CanWithdraw = application.AthleteUserID == actorUserID && canWithdraw(application, s.now())
	return application, nil
}

func (s *Service) Withdraw(ctx context.Context, actorUserID, applicationID int64) (Application, error) {
	application, err := s.GetAccessible(ctx, actorUserID, applicationID)
	if err != nil {
		return Application{}, err
	}
	if application.AthleteUserID != actorUserID {
		return Application{}, ErrForbidden
	}
	if application.Status != "pending" {
		return Application{}, ErrAlreadyReviewed
	}
	if s.now().After(application.RegistrationEndsAt) {
		return Application{}, ErrRegistrationClosed
	}
	result, err := s.repository.Withdraw(ctx, applicationID, actorUserID)
	if err != nil {
		return Application{}, err
	}
	result.CanWithdraw = false
	return result, nil
}

func (s *Service) ListForReview(ctx context.Context, filter ListFilter) ([]Application, error) {
	if filter.Limit < 1 || filter.Limit > MaxLimit || filter.Offset < 0 || (filter.EventID != nil && *filter.EventID <= 0) {
		return nil, ErrInvalidInput
	}
	if filter.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*filter.Status))
		if !validStatus(status) {
			return nil, ErrInvalidInput
		}
		filter.Status = &status
	}
	return s.repository.ListForReview(ctx, filter)
}

func (s *Service) GetForReview(ctx context.Context, applicationID int64) (Application, error) {
	if applicationID <= 0 {
		return Application{}, ErrInvalidInput
	}
	return s.repository.Get(ctx, applicationID)
}

func (s *Service) Review(ctx context.Context, reviewerUserID, applicationID int64, status string) (Application, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if reviewerUserID <= 0 || applicationID <= 0 || (status != "approved" && status != "rejected") {
		return Application{}, ErrInvalidInput
	}
	return s.repository.Review(ctx, applicationID, reviewerUserID, status)
}

func canWithdraw(application Application, now time.Time) bool {
	return application.Status == "pending" && !now.After(application.RegistrationEndsAt)
}

func validStatus(status string) bool {
	return status == "pending" || status == "approved" || status == "rejected" || status == "withdrawn"
}
