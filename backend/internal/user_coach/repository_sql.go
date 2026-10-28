package usercoach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	dbsqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/orenfsp/orencode-zenko.git/internal/user_coach/errs"
)

type SQLRepository struct {
	queries *dbsqlc.Queries
}

func NewSQLRepository(queries *dbsqlc.Queries) *SQLRepository {
	return &SQLRepository{queries: queries}
}

func (r *SQLRepository) SearchParticipants(ctx context.Context, coachID int64, search string, disciplineID *int64, limit, offset int32) ([]Participant, int64, error) {
	discipline := pgtype.Int8{}
	if disciplineID != nil {
		discipline = pgtype.Int8{Int64: *disciplineID, Valid: true}
	}
	total, err := r.queries.CountCoachParticipants(ctx, dbsqlc.CountCoachParticipantsParams{CoachID: coachID, Search: search, DisciplineID: discipline})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.SearchCoachParticipants(ctx, dbsqlc.SearchCoachParticipantsParams{CoachID: coachID, Search: search, DisciplineID: discipline, PageLimit: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, err
	}
	items := make([]Participant, 0, len(rows))
	for _, row := range rows {
		item := Participant{ID: row.ID, FirstName: row.FirstName, LastName: row.LastName,
			MiddleName: nullableText(row.MiddleName), Organization: nullableText(row.Organization), Municipality: nullableText(row.Municipality),
			Disciplines: []ParticipantDiscipline{}, CoachRelations: []ParticipantRelation{}}
		for _, field := range []struct {
			value  any
			target any
		}{{row.Disciplines, &item.Disciplines}, {row.CoachRelations, &item.CoachRelations}} {
			data, err := json.Marshal(field.value)
			if err != nil {
				return nil, 0, err
			}
			if err := json.Unmarshal(data, field.target); err != nil {
				return nil, 0, err
			}
		}
		items = append(items, item)
	}
	return items, total, nil
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func (r *SQLRepository) Create(
	ctx context.Context,
	participantUserID int64,
	coachUserID int64,
	disciplineID int64,
) (UserCoach, error) {
	row, err := r.queries.CreateUserCoach(
		ctx,
		dbsqlc.CreateUserCoachParams{
			ParticipantUserID: participantUserID,
			CoachUserID:       coachUserID,
			DisciplineID:      disciplineID,
		},
	)
	if err != nil {
		return UserCoach{}, mapRepositoryError(err)
	}

	return fromDB(row), nil
}

func (r *SQLRepository) ListByParticipant(
	ctx context.Context,
	participantUserID int64,
	limit int32,
	offset int32,
) ([]UserCoach, error) {
	rows, err := r.queries.ListUserCoachesByParticipant(
		ctx,
		dbsqlc.ListUserCoachesByParticipantParams{
			ParticipantUserID: participantUserID,
			PageLimit:         limit,
			PageOffset:        offset,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list coaches by participant: %w", err)
	}

	result := make([]UserCoach, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromDB(row))
	}

	return result, nil
}

func (r *SQLRepository) ListByCoach(
	ctx context.Context,
	coachUserID int64,
	limit int32,
	offset int32,
) ([]UserCoach, error) {
	rows, err := r.queries.ListUserCoachesByCoach(
		ctx,
		dbsqlc.ListUserCoachesByCoachParams{
			CoachUserID: coachUserID,
			PageLimit:   limit,
			PageOffset:  offset,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list participants by coach: %w", err)
	}

	result := make([]UserCoach, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromDB(row))
	}

	return result, nil
}

func (r *SQLRepository) Confirm(
	ctx context.Context,
	relationID int64,
	participantUserID int64,
) (UserCoach, error) {
	row, err := r.queries.ConfirmUserCoach(
		ctx,
		dbsqlc.ConfirmUserCoachParams{
			RelationID:        relationID,
			ParticipantUserID: participantUserID,
		},
	)
	if err != nil {
		return UserCoach{}, mapRepositoryError(err)
	}

	return fromDB(row), nil
}

func (r *SQLRepository) End(
	ctx context.Context,
	relationID int64,
	actorUserID int64,
) (UserCoach, error) {
	row, err := r.queries.EndUserCoach(
		ctx,
		dbsqlc.EndUserCoachParams{
			RelationID:  relationID,
			ActorUserID: actorUserID,
		},
	)
	if err != nil {
		return UserCoach{}, mapRepositoryError(err)
	}

	return fromDB(row), nil
}

func fromDB(row dbsqlc.UserCoach) UserCoach {
	return UserCoach{
		ID:           row.ID,
		UserID:       row.UserID,
		CoachUserID:  row.CoachUserID,
		DisciplineID: row.DisciplineID,
		Status:       Status(row.Status),
		RequestedAt:  row.RequestedAt.Time,
		ConfirmedAt:  nullableTime(row.ConfirmedAt),
		EndedAt:      nullableTime(row.EndedAt),
	}
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}

func mapRepositoryError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errs.ErrConflict
	}

	return fmt.Errorf("user coach repository: %w", err)
}
