package parentalconsent

import (
	"context"
	dbsqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository struct{ queries *dbsqlc.Queries }

func NewRepository(q *dbsqlc.Queries) *Repository { return &Repository{queries: q} }

func (r *Repository) List(ctx context.Context, status *string, limit, offset int32) ([]ConsentItem, error) {
	var pgStatus pgtype.Text
	if status != nil {
		pgStatus = pgtype.Text{String: *status, Valid: true}
	}

	rows, err := r.queries.ListParentalConsents(ctx, dbsqlc.ListParentalConsentsParams{
		Status:  pgStatus,
		PLimit:  limit,
		POffset: offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]ConsentItem, len(rows))
	for i, row := range rows {
		items[i] = ConsentItem{
			ID: row.ID, UserID: row.UserID, FirstName: row.FirstName, LastName: row.LastName,
			Email: row.Email, Status: row.Status, SubmittedAt: row.SubmittedAt.Time,
			ReviewedAt: &row.ReviewedAt.Time, FileID: &row.FileID.Int64,
		}
	}
	return items, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.queries.UpdateParentalConsentStatus(ctx, dbsqlc.UpdateParentalConsentStatusParams{
		ID: id, Status: status,
	})
}