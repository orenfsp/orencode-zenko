package usercoach

import "context"

type Repository interface {
	SearchParticipants(ctx context.Context, coachID int64, search string, disciplineID *int64, limit, offset int32) ([]Participant, int64, error)
	Create(
		ctx context.Context,
		participantUserID int64,
		coachUserID int64,
		disciplineID int64,
	) (UserCoach, error)

	ListByParticipant(
		ctx context.Context,
		participantUserID int64,
		limit int32,
		offset int32,
	) ([]UserCoach, error)

	ListByCoach(
		ctx context.Context,
		coachUserID int64,
		limit int32,
		offset int32,
	) ([]UserCoach, error)

	Confirm(
		ctx context.Context,
		relationID int64,
		participantUserID int64,
	) (UserCoach, error)

	End(
		ctx context.Context,
		relationID int64,
		actorUserID int64,
	) (UserCoach, error)
}
