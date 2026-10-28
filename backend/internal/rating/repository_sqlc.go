package rating

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
	pool    interface {
		BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	}
}

func NewRepository(queries *db.Queries, pool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) Repository {
	return &sqlRepository{queries: queries, pool: pool}
}

func (r *sqlRepository) List(ctx context.Context, filter Filter) ([]Entry, error) {
	rows, err := r.queries.ListAthleteRating(ctx, db.ListAthleteRatingParams{
		Year: filter.Year, DisciplineID: optionalInt(filter.DisciplineID), MunicipalityID: optionalInt(filter.MunicipalityID),
		AgeGroupID: optionalInt(filter.AgeGroupID), PLimit: filter.Limit, POffset: filter.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]Entry, 0, len(rows))
	for _, row := range rows {
		items = append(items, Entry{Position: row.Position, UserID: row.UserID, LastName: row.LastName, FirstName: row.FirstName,
			MiddleName: textPointer(row.MiddleName), MunicipalityName: textPointer(row.MunicipalityName), TotalPoints: row.TotalPoints,
			ResultCount: row.ResultCount, Wins: row.Wins})
	}
	return items, nil
}

func (r *sqlRepository) DisciplineExists(ctx context.Context, disciplineID int64) (bool, error) {
	return r.queries.RatingDisciplineExists(ctx, disciplineID)
}

func (r *sqlRepository) AgeGroupExists(ctx context.Context, ageGroupID int64) (bool, error) {
	return r.queries.RatingAgeGroupExists(ctx, ageGroupID)
}

func (r *sqlRepository) UserHasRole(ctx context.Context, userID int64, roleCode string) (bool, error) {
	return r.queries.RatingUserHasRole(ctx, db.RatingUserHasRoleParams{UserID: userID, RoleCode: roleCode})
}

func (r *sqlRepository) GetSelection(ctx context.Context, scope SelectionScope) (Selection, error) {
	id, err := r.queries.GetRatingSelectionIDByScope(ctx, db.GetRatingSelectionIDByScopeParams{
		SeasonYear: int16(scope.SeasonYear), DisciplineID: scope.DisciplineID, AgeGroupID: scope.AgeGroupID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Selection{}, ErrSelectionNotFound
	}
	if err != nil {
		return Selection{}, err
	}
	return r.loadSelection(ctx, r.queries, id)
}

func (r *sqlRepository) SaveSelection(ctx context.Context, actorID int64, request SaveSelectionRequest) (Selection, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Selection{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	id, err := queries.UpsertRatingSelection(ctx, db.UpsertRatingSelectionParams{
		SeasonYear: int16(request.SeasonYear), DisciplineID: request.DisciplineID, AgeGroupID: request.AgeGroupID,
		CaptainUserID: optionalInt(request.CaptainUserID), CoachUserID: optionalInt(request.CoachUserID), CreatedByUserID: actorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Selection{}, ErrSelectionLocked
	}
	if err != nil {
		return Selection{}, err
	}
	if err := queries.DeleteRatingSelectionMembers(ctx, id); err != nil {
		return Selection{}, err
	}
	for index, userID := range request.MemberIDs {
		if err := queries.AddRatingSelectionMember(ctx, db.AddRatingSelectionMemberParams{
			SelectionID: id, UserID: userID, SelectionPosition: int32(index + 1),
		}); err != nil {
			return Selection{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Selection{}, err
	}
	return r.loadSelection(ctx, r.queries, id)
}

func (r *sqlRepository) LockSelection(ctx context.Context, selectionID int64) (Selection, error) {
	_, err := r.queries.LockRatingSelection(ctx, selectionID)
	if errors.Is(err, pgx.ErrNoRows) {
		_, statusErr := r.queries.GetRatingSelectionStatus(ctx, selectionID)
		if errors.Is(statusErr, pgx.ErrNoRows) {
			return Selection{}, ErrSelectionNotFound
		}
		if statusErr != nil {
			return Selection{}, statusErr
		}
		return Selection{}, ErrSelectionLocked
	}
	if err != nil {
		return Selection{}, err
	}
	return r.loadSelection(ctx, r.queries, selectionID)
}

type selectionQueries interface {
	GetRatingSelection(context.Context, int64) (db.GetRatingSelectionRow, error)
	ListRatingSelectionMembers(context.Context, int64) ([]db.ListRatingSelectionMembersRow, error)
}

func (r *sqlRepository) loadSelection(ctx context.Context, queries selectionQueries, id int64) (Selection, error) {
	row, err := queries.GetRatingSelection(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Selection{}, ErrSelectionNotFound
	}
	if err != nil {
		return Selection{}, err
	}
	members, err := queries.ListRatingSelectionMembers(ctx, id)
	if err != nil {
		return Selection{}, err
	}
	result := Selection{
		ID: row.ID, SeasonYear: int32(row.SeasonYear),
		Discipline: Reference{ID: row.DisciplineID, Name: row.DisciplineName},
		AgeGroup:   Reference{ID: row.AgeGroupID, Name: row.AgeGroupName},
		Status:     row.Status, CaptainUserID: intPointer(row.CaptainUserID), CoachUserID: intPointer(row.CoachUserID),
		Members: make([]SelectionMember, 0, len(members)), CreatedByUserID: row.CreatedByUserID,
		LockedAt: timePointer(row.LockedAt), CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
	for _, member := range members {
		result.Members = append(result.Members, SelectionMember{
			SelectionPosition: member.SelectionPosition, UserID: member.UserID,
			LastName: member.LastName, FirstName: member.FirstName, MiddleName: textPointer(member.MiddleName),
		})
	}
	return result, nil
}

func optionalInt(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}
func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func intPointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
