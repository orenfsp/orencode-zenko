package rank

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) CreateRank(ctx context.Context, models Create) (Response, error) {
	rows, err := r.queries.CreateRank(ctx, db.CreateRankParams{
		Name: models.Name,
		Ttl: pgtype.Int4{
			Int32: int32(models.TTL),
			Valid: true,
		},
	})
	if err != nil {
		return Response{}, err
	}
	return Response{
		ID:        int32(rows.ID),
		Name:      rows.Name,
		TTL:       int(rows.Ttl.Int32),
		CreatedAt: rows.CreatedAt.Time,
		UpdatedAt: rows.UpdatedAt.Time,
	}, nil
}

func (r *sqlRepository) DeleteExpired(ctx context.Context) (int64, error) {
	legacyCount, err := r.queries.DeleteExpiredUserRanks(ctx)
	if err != nil {
		return 0, err
	}
	disciplineCount, err := r.queries.DeleteExpiredUserRankDisciplines(ctx)
	if err != nil {
		return 0, err
	}
	return legacyCount + disciplineCount, nil
}

func (r *sqlRepository) GetRank(ctx context.Context, id int32) (Response, error) {
	rows, err := r.queries.GetRank(ctx, int64(id))
	if err != nil {
		return Response{}, err
	}

	return Response{
		ID:        int32(rows.ID),
		Name:      rows.Name,
		TTL:       int(rows.Ttl.Int32),
		CreatedAt: rows.CreatedAt.Time,
		UpdatedAt: rows.UpdatedAt.Time,
	}, nil
}
func (r *sqlRepository) GetListRank(ctx context.Context) ([]*Response, error) {
	rows, err := r.queries.GetListRank(ctx)
	if err != nil {
		return nil, err
	}

	arry := make([]*Response, 0, len(rows))

	for _, i := range rows {
		arry = append(arry, &Response{
			ID:        int32(i.ID),
			Name:      i.Name,
			TTL:       int(i.Ttl.Int32),
			CreatedAt: i.CreatedAt.Time,
			UpdatedAt: i.UpdatedAt.Time,
		})
	}
	return arry, nil
}
