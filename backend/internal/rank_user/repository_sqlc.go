package rankuser

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) HasConfirmedCoach(ctx context.Context, athleteUserID, disciplineID int64) (bool, error) {
	return r.queries.HasConfirmedCoachForRankApplication(ctx, db.HasConfirmedCoachForRankApplicationParams{
		AthleteUserID: athleteUserID,
		DisciplineID:  disciplineID,
	})
}

func (r *sqlRepository) IsConfirmedCoach(ctx context.Context, athleteUserID, coachUserID, disciplineID int64) (bool, error) {
	return r.queries.IsConfirmedCoachForRankApplication(ctx, db.IsConfirmedCoachForRankApplicationParams{
		AthleteUserID: athleteUserID,
		CoachUserID:   coachUserID,
		DisciplineID:  disciplineID,
	})
}

func (r *sqlRepository) Create(ctx context.Context, athleteUserID, rankID, disciplineID, submittedByUserID int64) (Application, error) {
	row, err := r.queries.CreateRankApplication(ctx, db.CreateRankApplicationParams{
		AthleteUserID:     athleteUserID,
		RankID:            rankID,
		DisciplineID:      disciplineID,
		SubmittedByUserID: submittedByUserID,
	})
	if err != nil {
		return Application{}, mapRepositoryError(err)
	}
	return fromDB(row), nil
}

func (r *sqlRepository) Review(ctx context.Context, applicationID, reviewerUserID int64, status string, rejectionReason *string) (Application, error) {
	row, err := r.queries.ReviewRankApplication(ctx, db.ReviewRankApplicationParams{
		ID:               applicationID,
		Status:           status,
		ReviewedByUserID: pgtype.Int8{Int64: reviewerUserID, Valid: true},
		RejectionReason:  optionalText(rejectionReason),
	})
	if err != nil {
		return Application{}, mapRepositoryError(err)
	}
	return fromDB(row), nil
}

func (r *sqlRepository) ListByAthlete(ctx context.Context, athleteUserID int64) ([]Application, error) {
	rows, err := r.queries.ListRankApplicationsByAthlete(ctx, athleteUserID)
	if err != nil {
		return nil, fmt.Errorf("list rank applications: %w", err)
	}
	result := make([]Application, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromDB(row))
	}
	return result, nil
}

func fromDB(row db.RankApplication) Application {
	return Application{
		ID:                row.ID,
		AthleteUserID:     row.AthleteUserID,
		RankID:            row.RankID,
		DisciplineID:      row.DisciplineID,
		SubmittedByUserID: row.SubmittedByUserID,
		Status:            row.Status,
		ReviewedByUserID:  optionalInt64(row.ReviewedByUserID),
		ReviewedAt:        optionalTime(row.ReviewedAt),
		RejectionReason:   optionalString(row.RejectionReason),
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
	}
}

func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalInt64(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func optionalTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func optionalString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func mapRepositoryError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505":
		return ErrConflict
	case "23503":
		return ErrInvalidInput
	default:
		return err
	}
}
