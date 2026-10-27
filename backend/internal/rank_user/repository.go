package rankuser

import "context"

type Repository interface {
	HasConfirmedCoach(ctx context.Context, athleteUserID, disciplineID int64) (bool, error)
	IsConfirmedCoach(ctx context.Context, athleteUserID, coachUserID, disciplineID int64) (bool, error)
	Create(ctx context.Context, athleteUserID, rankID, disciplineID, submittedByUserID int64) (Application, error)
	Review(ctx context.Context, applicationID, reviewerUserID int64, status string, rejectionReason *string) (Application, error)
	ListByAthlete(ctx context.Context, athleteUserID int64) ([]Application, error)
}
