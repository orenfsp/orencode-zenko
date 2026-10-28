package rankuser

import (
	"context"
	"time"
)

type Repository interface {
	HasConfirmedCoach(ctx context.Context, athleteUserID, disciplineID int64) (bool, error)
	IsConfirmedCoach(ctx context.Context, athleteUserID, coachUserID, disciplineID int64) (bool, error)
	ValidateBasis(ctx context.Context, athleteUserID, disciplineID, competitionID, resultID int64) (BasisValidation, error)
	ValidateEVSK(ctx context.Context, applicationID int64) (bool, error)
	Create(ctx context.Context, athleteUserID, rankID, disciplineID, submittedByUserID int64, applicationType string, competitionID, resultID int64) (Application, error)
	Review(ctx context.Context, applicationID, reviewerUserID int64, status string, rejectionReason *string, assignedAt, expiresAt *time.Time, orderNumber *string) (Application, error)
	ListByAthlete(ctx context.Context, athleteUserID int64) ([]Application, error)
	ListPending(ctx context.Context, limit, offset int32) ([]Application, error)
	ListHistory(ctx context.Context, athleteUserID int64) ([]RankHistory, error)
}
