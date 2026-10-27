package municipality

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	municipalityerrs "github.com/orenfsp/orencode-zenko.git/internal/municipality/errs"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) Create(ctx context.Context, input CreateMunicipality) (Municipality, error) {
	row, err := r.queries.CreateMunicipality(ctx, input.Name)
	if err != nil {
		return Municipality{}, mapCreateError(err)
	}
	return Municipality{ID: row.ID, Name: row.Name}, nil
}

func (r *sqlRepository) GetByID(ctx context.Context, id int64) (Municipality, error) {
	row, err := r.queries.GetMunicipalityByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Municipality{}, municipalityerrs.ErrNotFound
	}
	if err != nil {
		return Municipality{}, err
	}
	return Municipality{ID: row.ID, Name: row.Name}, nil
}

func (r *sqlRepository) List(ctx context.Context, limit, offset int32) ([]Municipality, error) {
	rows, err := r.queries.ListMunicipalities(ctx, db.ListMunicipalitiesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	result := make([]Municipality, 0, len(rows))
	for _, row := range rows {
		result = append(result, Municipality{ID: row.ID, Name: row.Name})
	}
	return result, nil
}

func mapCreateError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	if pgErr.Code == "23505" && pgErr.ConstraintName == "municipality_name_key" {
		return municipalityerrs.ErrNameConflict
	}
	return err
}
