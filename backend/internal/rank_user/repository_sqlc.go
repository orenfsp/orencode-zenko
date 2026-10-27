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

func (r *sqlRepository) ValidateBasis(ctx context.Context, athleteUserID, disciplineID, competitionID, resultID int64) (BasisValidation, error) {
	row, err := r.queries.ValidateRankApplicationBasis(ctx, db.ValidateRankApplicationBasisParams{
		ResultID: resultID, AthleteUserID: athleteUserID, DisciplineID: disciplineID, CompetitionID: competitionID,
	})
	if err != nil {
		return BasisValidation{}, err
	}
	return BasisValidation{
		ResultExists: row.ResultExists, AthleteMatches: row.AthleteMatches,
		DisciplineMatches: row.DisciplineMatches, CompetitionMatches: row.CompetitionMatches,
		ResultFinished: row.ResultFinished,
	}, nil
}

func (r *sqlRepository) Create(ctx context.Context, athleteUserID, rankID, disciplineID, submittedByUserID int64, applicationType string, competitionID, resultID int64) (Application, error) {
	row, err := r.queries.CreateRankApplication(ctx, db.CreateRankApplicationParams{
		AthleteUserID:     athleteUserID,
		RankID:            rankID,
		DisciplineID:      disciplineID,
		SubmittedByUserID: submittedByUserID,
		ApplicationType:   applicationType,
		CompetitionID:     pgtype.Int8{Int64: competitionID, Valid: true},
		ResultID:          pgtype.Int8{Int64: resultID, Valid: true},
	})
	if err != nil {
		return Application{}, mapRepositoryError(err)
	}
	return fromDB(row), nil
}

func (r *sqlRepository) Review(ctx context.Context, applicationID, reviewerUserID int64, status string, rejectionReason *string, assignedAt, expiresAt *time.Time, orderNumber *string) (Application, error) {
	row, err := r.queries.ReviewRankApplication(ctx, db.ReviewRankApplicationParams{
		ID:               applicationID,
		Status:           status,
		ReviewedByUserID: pgtype.Int8{Int64: reviewerUserID, Valid: true},
		RejectionReason:  optionalText(rejectionReason),
		AssignedAt:       optionalDate(assignedAt),
		ExpiresAt:        optionalDate(expiresAt),
		OrderNumber:      optionalText(orderNumber),
	})
	if err != nil {
		return Application{}, mapRepositoryError(err)
	}
	return fromDB(row), nil
}

func (r *sqlRepository) ListPending(ctx context.Context, limit, offset int32) ([]Application, error) {
	rows, err := r.queries.ListPendingRankApplications(ctx, db.ListPendingRankApplicationsParams{POffset: offset, PLimit: limit})
	if err != nil {
		return nil, err
	}
	result := make([]Application, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromDB(row))
	}
	return result, nil
}

func (r *sqlRepository) ListHistory(ctx context.Context, athleteUserID int64) ([]RankHistory, error) {
	rows, err := r.queries.ListUserRankHistory(ctx, athleteUserID)
	if err != nil {
		return nil, err
	}
	result := make([]RankHistory, 0, len(rows))
	for _, row := range rows {
		result = append(result, RankHistory{
			ID: row.ID, UserID: row.UserID, RankID: row.RankID, RankName: row.RankName,
			DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
			ResultID: optionalInt64(row.ResultID), AwardedAt: optionalDateValue(row.AwardedAt),
			ExpiresAt: optionalDateValue(row.ExpiresAt), OrderNumber: optionalString(row.OrderNumber),
			CompetitionID: optionalInt64(row.CompetitionID), CompetitionName: optionalString(row.CompetitionName),
		})
	}
	return result, nil
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
		ApplicationType:   row.ApplicationType,
		CompetitionID:     optionalInt64(row.CompetitionID),
		ResultID:          optionalInt64(row.ResultID),
		SubmittedByUserID: row.SubmittedByUserID,
		Status:            row.Status,
		ReviewedByUserID:  optionalInt64(row.ReviewedByUserID),
		ReviewedAt:        optionalTime(row.ReviewedAt),
		RejectionReason:   optionalString(row.RejectionReason),
		AssignedAt:        optionalDateValue(row.AssignedAt),
		ExpiresAt:         optionalDateValue(row.ExpiresAt),
		OrderNumber:       optionalString(row.OrderNumber),
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
	}
}

func optionalDate(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *value, Valid: true}
}

func optionalDateValue(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
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
