package competitionapplication

import "context"

type Repository interface {
	GetEventContext(ctx context.Context, eventID int64) (EventContext, error)
	IsConfirmedCoach(ctx context.Context, athleteUserID, coachUserID, disciplineID int64) (bool, error)
	Create(ctx context.Context, eventID, athleteUserID, submittedByUserID int64) (Application, error)
	Get(ctx context.Context, applicationID int64) (Application, error)
	ListAccessible(ctx context.Context, actorUserID int64) ([]Application, error)
	ListForReview(ctx context.Context, filter ListFilter) ([]Application, error)
	Review(ctx context.Context, applicationID, reviewerUserID int64, status string) (Application, error)
	Withdraw(ctx context.Context, applicationID, athleteUserID int64) (Application, error)
}
