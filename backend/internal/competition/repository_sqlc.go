package competition

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) ListPublished(ctx context.Context, limit, offset int32) ([]Competition, error) {
	rows, err := r.queries.ListPublishedCompetitions(ctx, db.ListPublishedCompetitionsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]Competition, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromListRow(row))
	}
	return result, nil
}

func (r *sqlRepository) GetPublished(ctx context.Context, id int64) (Competition, error) {
	row, err := r.queries.GetPublishedCompetition(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Competition{}, competitionerrs.ErrNotFound
		}
		return Competition{}, err
	}
	return fromGetRow(row), nil
}

func fromListRow(row db.ListPublishedCompetitionsRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description),
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func fromGetRow(row db.GetPublishedCompetitionRow) Competition {
	return Competition{
		ID: row.ID, Name: row.Name, Description: textPointer(row.Description),
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		CompetitionTypeID: row.CompetitionTypeID, CompetitionTypeName: row.CompetitionTypeName,
		CompetitionLevelID: row.CompetitionLevelID, CompetitionLevelCode: textPointer(row.CompetitionLevelCode),
		CompetitionLevelName: textPointer(row.CompetitionLevelName), Format: row.Format,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		RegistrationStartsAt: row.RegistrationStartsAt.Time, RegistrationEndsAt: row.RegistrationEndsAt.Time,
		VenueAddress: textPointer(row.VenueAddress), OnlineURL: textPointer(row.OnlineUrl), Status: row.Status,
		CalendarPlanID: int64Pointer(row.CalendarPlanID), CalendarPlanName: textPointer(row.CalendarPlanName),
		CalendarPlanYear: int16Pointer(row.CalendarPlanYear), CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		RegulationURL: textPointer(row.RegulationUrl),
	}
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func int16Pointer(value pgtype.Int2) *int16 {
	if !value.Valid {
		return nil
	}
	return &value.Int16
}
