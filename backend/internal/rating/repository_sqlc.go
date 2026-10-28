package rating

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct{ queries *db.Queries }

func NewRepository(queries *db.Queries) Repository { return &sqlRepository{queries: queries} }

func (r *sqlRepository) List(ctx context.Context, filter Filter) ([]Entry, error) {
	rows, err := r.queries.ListAthleteRating(ctx, db.ListAthleteRatingParams{
		Year: filter.Year, DisciplineID: optionalInt(filter.DisciplineID), MunicipalityID: optionalInt(filter.MunicipalityID),
		PLimit: filter.Limit, POffset: filter.Offset,
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
