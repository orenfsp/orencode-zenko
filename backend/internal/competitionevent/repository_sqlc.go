package competitionevent

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *db.Queries, pool *pgxpool.Pool) Repository {
	return &sqlRepository{queries: queries, pool: pool}
}

func (r *sqlRepository) ListPublic(ctx context.Context, competitionID int64) ([]Event, error) {
	rows, err := r.queries.ListPublicCompetitionEvents(ctx, db.ListPublicCompetitionEventsParams{
		CompetitionID: competitionID, PublicStatuses: publicCompetitionStatuses,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		exists, err := r.queries.PublicCompetitionForEventsExists(ctx, db.PublicCompetitionForEventsExistsParams{
			CompetitionID: competitionID, PublicStatuses: publicCompetitionStatuses,
		})
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrCompetitionMissing
		}
	}
	items := make([]Event, 0, len(rows))
	for _, row := range rows {
		items = append(items, eventFromValues(row.ID, row.CompetitionID, row.DisciplineID,
			row.DisciplineName, row.AgeGroupID, row.AgeGroupName, row.ProtocolUrl))
	}
	return items, nil
}

func (r *sqlRepository) GetPublic(ctx context.Context, competitionID, eventID int64) (Event, error) {
	row, err := r.queries.GetPublicCompetitionEvent(ctx, db.GetPublicCompetitionEventParams{
		CompetitionID: competitionID, EventID: eventID, PublicStatuses: publicCompetitionStatuses,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, ErrEventMissing
	}
	if err != nil {
		return Event{}, err
	}
	return eventFromValues(row.ID, row.CompetitionID, row.DisciplineID, row.DisciplineName,
		row.AgeGroupID, row.AgeGroupName, row.ProtocolUrl), nil
}

func (r *sqlRepository) ListAdmin(ctx context.Context, competitionID int64) ([]Event, error) {
	rows, err := r.queries.ListCompetitionEventsForAdmin(ctx, competitionID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		exists, err := r.queries.CompetitionForEventsExists(ctx, competitionID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrCompetitionMissing
		}
	}
	items := make([]Event, 0, len(rows))
	for _, row := range rows {
		items = append(items, eventFromValues(row.ID, row.CompetitionID, row.DisciplineID,
			row.DisciplineName, row.AgeGroupID, row.AgeGroupName, row.ProtocolUrl))
	}
	return items, nil
}

func (r *sqlRepository) GetAdmin(ctx context.Context, competitionID, eventID int64) (Event, error) {
	row, err := r.queries.GetCompetitionEventForAdmin(ctx, db.GetCompetitionEventForAdminParams{
		CompetitionID: competitionID, EventID: eventID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, ErrEventMissing
	}
	if err != nil {
		return Event{}, err
	}
	return eventFromValues(row.ID, row.CompetitionID, row.DisciplineID, row.DisciplineName,
		row.AgeGroupID, row.AgeGroupName, row.ProtocolUrl), nil
}

func (r *sqlRepository) CreateDraft(ctx context.Context, competitionID int64, ageGroupID *int64) (Event, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Event{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	parent, err := lockDraftCompetition(ctx, queries, competitionID)
	if err != nil {
		return Event{}, err
	}
	if err := validateAgeGroup(ctx, queries, ageGroupID); err != nil {
		return Event{}, err
	}
	row, err := queries.CreateCompetitionEvent(ctx, db.CreateCompetitionEventParams{
		CompetitionID: competitionID, DisciplineID: parent.DisciplineID, AgeGroupID: nullableInt64(ageGroupID),
	})
	if err != nil {
		return Event{}, mapWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Event{}, err
	}
	return eventFromValues(row.ID, row.CompetitionID, row.DisciplineID, row.DisciplineName,
		row.AgeGroupID, row.AgeGroupName, row.ProtocolUrl), nil
}

func (r *sqlRepository) UpdateDraft(ctx context.Context, competitionID, eventID int64, ageGroupID *int64) (Event, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Event{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	if _, err := lockDraftCompetition(ctx, queries, competitionID); err != nil {
		return Event{}, err
	}
	current, err := queries.LockCompetitionEventForMutation(ctx, db.LockCompetitionEventForMutationParams{
		CompetitionID: competitionID, EventID: eventID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, ErrEventMissing
	}
	if err != nil {
		return Event{}, err
	}
	if sameNullableID(int64Pointer(current.AgeGroupID), ageGroupID) {
		if err := tx.Commit(ctx); err != nil {
			return Event{}, err
		}
		return eventFromValues(current.ID, current.CompetitionID, current.DisciplineID,
			current.DisciplineName, current.AgeGroupID, current.AgeGroupName, current.ProtocolUrl), nil
	}
	if err := validateAgeGroup(ctx, queries, ageGroupID); err != nil {
		return Event{}, err
	}
	hasApplications, err := queries.CompetitionEventHasApplications(ctx, eventID)
	if err != nil {
		return Event{}, err
	}
	if hasApplications {
		return Event{}, ErrHasApplications
	}
	row, err := queries.UpdateCompetitionEvent(ctx, db.UpdateCompetitionEventParams{
		CompetitionID: competitionID, EventID: eventID, AgeGroupID: nullableInt64(ageGroupID),
	})
	if err != nil {
		return Event{}, mapWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Event{}, err
	}
	return eventFromValues(row.ID, row.CompetitionID, row.DisciplineID, row.DisciplineName,
		row.AgeGroupID, row.AgeGroupName, row.ProtocolUrl), nil
}

func (r *sqlRepository) DeleteDraft(ctx context.Context, competitionID, eventID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	if _, err := lockDraftCompetition(ctx, queries, competitionID); err != nil {
		return err
	}
	if _, err := queries.LockCompetitionEventForMutation(ctx, db.LockCompetitionEventForMutationParams{
		CompetitionID: competitionID, EventID: eventID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return ErrEventMissing
	} else if err != nil {
		return err
	}
	hasApplications, err := queries.CompetitionEventHasApplications(ctx, eventID)
	if err != nil {
		return err
	}
	if hasApplications {
		return ErrHasApplications
	}
	rows, err := queries.DeleteCompetitionEvent(ctx, db.DeleteCompetitionEventParams{
		CompetitionID: competitionID, EventID: eventID,
	})
	if err != nil {
		return mapWriteError(err)
	}
	if rows != 1 {
		return ErrEventMissing
	}
	return tx.Commit(ctx)
}

func lockDraftCompetition(ctx context.Context, queries *db.Queries, competitionID int64) (db.LockCompetitionForEventMutationRow, error) {
	parent, err := queries.LockCompetitionForEventMutation(ctx, competitionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.LockCompetitionForEventMutationRow{}, ErrCompetitionMissing
	}
	if err != nil {
		return db.LockCompetitionForEventMutationRow{}, err
	}
	if parent.Status != "draft" {
		return db.LockCompetitionForEventMutationRow{}, ErrNotDraft
	}
	return parent, nil
}

func validateAgeGroup(ctx context.Context, queries *db.Queries, ageGroupID *int64) error {
	if ageGroupID == nil {
		return nil
	}
	exists, err := queries.AgeGroupExists(ctx, *ageGroupID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAgeGroupMissing
	}
	return nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return err
	}
	switch pgErr.ConstraintName {
	case "competition_event_age_group_id_fkey":
		return ErrAgeGroupMissing
	case "competition_application_event_id_fkey":
		return ErrHasApplications
	case "competition_event_competition_id_fkey":
		return ErrCompetitionMissing
	default:
		return err
	}
}

func nullableInt64(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func int64Pointer(value pgtype.Int8) *int64 {
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

func sameNullableID(left, right *int64) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func eventFromValues(id, competitionID, disciplineID int64, disciplineName string,
	ageGroupID pgtype.Int8, ageGroupName, protocolURL pgtype.Text,
) Event {
	return Event{
		ID: id, CompetitionID: competitionID, DisciplineID: disciplineID, DisciplineName: disciplineName,
		AgeGroupID: int64Pointer(ageGroupID), AgeGroupName: textPointer(ageGroupName),
		ProtocolURL: textPointer(protocolURL),
	}
}
