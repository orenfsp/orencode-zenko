package competitionapplication

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct{ queries *db.Queries }

func NewRepository(queries *db.Queries) Repository { return &sqlRepository{queries: queries} }

func (r *sqlRepository) GetEventContext(ctx context.Context, eventID int64) (EventContext, error) {
	row, err := r.queries.GetCompetitionApplicationEventContext(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return EventContext{}, ErrNotFound
	}
	if err != nil {
		return EventContext{}, err
	}
	return EventContext{EventID: row.EventID, DisciplineID: row.DisciplineID, CompetitionID: row.CompetitionID,
		CompetitionStatus: row.CompetitionStatus, RegistrationStartsAt: row.RegistrationStartsAt.Time,
		RegistrationEndsAt: row.RegistrationEndsAt.Time}, nil
}

func (r *sqlRepository) IsConfirmedCoach(ctx context.Context, athleteUserID, coachUserID, disciplineID int64) (bool, error) {
	return r.queries.IsConfirmedCoachForCompetitionApplication(ctx, db.IsConfirmedCoachForCompetitionApplicationParams{
		AthleteUserID: athleteUserID, CoachUserID: coachUserID, DisciplineID: disciplineID,
	})
}

func (r *sqlRepository) Create(ctx context.Context, eventID, athleteUserID, submittedByUserID int64) (Application, error) {
	row, err := r.queries.CreateCompetitionApplication(ctx, db.CreateCompetitionApplicationParams{
		EventID: eventID, AthleteUserID: athleteUserID, SubmittedByUserID: submittedByUserID,
	})
	if err != nil {
		return Application{}, mapRepositoryError(err)
	}
	return r.Get(ctx, row.ID)
}

func (r *sqlRepository) Get(ctx context.Context, applicationID int64) (Application, error) {
	row, err := r.queries.GetCompetitionApplication(ctx, applicationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, err
	}
	return mapGetRow(row), nil
}

func (r *sqlRepository) ListAccessible(ctx context.Context, actorUserID int64) ([]Application, error) {
	rows, err := r.queries.ListAccessibleCompetitionApplications(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	result := make([]Application, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAccessibleRow(row))
	}
	return result, nil
}

func (r *sqlRepository) ListForReview(ctx context.Context, filter ListFilter) ([]Application, error) {
	rows, err := r.queries.ListCompetitionApplicationsForReview(ctx, db.ListCompetitionApplicationsForReviewParams{
		PEventID: optionalInt8(filter.EventID), PStatus: optionalText(filter.Status), POffset: filter.Offset, PLimit: filter.Limit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]Application, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapReviewRow(row))
	}
	return result, nil
}

func (r *sqlRepository) Review(ctx context.Context, applicationID, reviewerUserID int64, status string) (Application, error) {
	row, err := r.queries.ReviewCompetitionApplication(ctx, db.ReviewCompetitionApplicationParams{
		Status: status, ReviewedByUserID: pgtype.Int8{Int64: reviewerUserID, Valid: true}, ApplicationID: applicationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrAlreadyReviewed
	}
	if err != nil {
		return Application{}, err
	}
	return r.Get(ctx, row.ID)
}

func (r *sqlRepository) Withdraw(ctx context.Context, applicationID, athleteUserID int64) (Application, error) {
	row, err := r.queries.WithdrawCompetitionApplication(ctx, db.WithdrawCompetitionApplicationParams{
		ApplicationID: applicationID, AthleteUserID: athleteUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrForbidden
	}
	if err != nil {
		return Application{}, err
	}
	return r.Get(ctx, row.ID)
}

func mapGetRow(row db.GetCompetitionApplicationRow) Application {
	return newApplication(row.ID, row.EventID, row.AthleteUserID, row.SubmittedByUserID, row.Status, row.SubmittedAt,
		row.ReviewedByUserID, row.ReviewedAt, row.CompetitionID, row.DisciplineID, row.DisciplineName,
		row.AgeGroupID, row.AgeGroupName, row.CompetitionName, row.CompetitionStartAt, row.CompetitionEndAt,
		row.RegistrationStartsAt, row.RegistrationEndsAt, row.AthleteFirstName, row.AthleteLastName,
		row.AthleteMiddleName, row.AthleteEmail, row.SubmitterFirstName, row.SubmitterLastName, row.SubmitterMiddleName)
}

func mapAccessibleRow(row db.ListAccessibleCompetitionApplicationsRow) Application {
	return newApplication(row.ID, row.EventID, row.AthleteUserID, row.SubmittedByUserID, row.Status, row.SubmittedAt,
		row.ReviewedByUserID, row.ReviewedAt, row.CompetitionID, row.DisciplineID, row.DisciplineName,
		row.AgeGroupID, row.AgeGroupName, row.CompetitionName, row.CompetitionStartAt, row.CompetitionEndAt,
		row.RegistrationStartsAt, row.RegistrationEndsAt, row.AthleteFirstName, row.AthleteLastName,
		row.AthleteMiddleName, row.AthleteEmail, row.SubmitterFirstName, row.SubmitterLastName, row.SubmitterMiddleName)
}

func mapReviewRow(row db.ListCompetitionApplicationsForReviewRow) Application {
	return newApplication(row.ID, row.EventID, row.AthleteUserID, row.SubmittedByUserID, row.Status, row.SubmittedAt,
		row.ReviewedByUserID, row.ReviewedAt, row.CompetitionID, row.DisciplineID, row.DisciplineName,
		row.AgeGroupID, row.AgeGroupName, row.CompetitionName, row.CompetitionStartAt, row.CompetitionEndAt,
		row.RegistrationStartsAt, row.RegistrationEndsAt, row.AthleteFirstName, row.AthleteLastName,
		row.AthleteMiddleName, row.AthleteEmail, row.SubmitterFirstName, row.SubmitterLastName, row.SubmitterMiddleName)
}

func newApplication(id, eventID, athleteID, submitterID int64, status string, submitted pgtype.Timestamptz,
	reviewer pgtype.Int8, reviewed pgtype.Timestamptz, competitionID, disciplineID int64, disciplineName string,
	ageGroupID pgtype.Int8, ageGroupName pgtype.Text, competitionName string, competitionStart, competitionEnd,
	registrationStart, registrationEnd pgtype.Timestamptz, athleteFirst, athleteLast string, athleteMiddle pgtype.Text,
	athleteEmail, submitterFirst, submitterLast string, submitterMiddle pgtype.Text) Application {
	return Application{ID: id, EventID: eventID, AthleteUserID: athleteID, SubmittedByUserID: submitterID,
		Status: status, SubmittedAt: submitted.Time, ReviewedByUserID: int8Pointer(reviewer), ReviewedAt: timePointer(reviewed),
		CompetitionID: competitionID, CompetitionName: competitionName, CompetitionStartAt: competitionStart.Time,
		CompetitionEndAt: competitionEnd.Time, RegistrationStartsAt: registrationStart.Time, RegistrationEndsAt: registrationEnd.Time,
		DisciplineID: disciplineID, DisciplineName: disciplineName, AgeGroupID: int8Pointer(ageGroupID), AgeGroupName: textPointer(ageGroupName),
		AthleteFirstName: athleteFirst, AthleteLastName: athleteLast, AthleteMiddleName: textPointer(athleteMiddle),
		AthleteEmail: athleteEmail, SubmitterFirstName: submitterFirst, SubmitterLastName: submitterLast,
		SubmitterMiddleName: textPointer(submitterMiddle)}
}

func optionalInt8(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}
func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}
func int8Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}
func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func mapRepositoryError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch {
	case pgErr.Code == "23505":
		return ErrConflict
	case pgErr.ConstraintName == "competition_application_valid_anti_doping_check" || pgErr.Code == "23514":
		return ErrAntiDopingRequired
	case pgErr.Code == "23503":
		return ErrNotFound
	default:
		return err
	}
}
