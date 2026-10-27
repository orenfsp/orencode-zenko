package reference

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

func (r *sqlRepository) ListDisciplines(ctx context.Context) ([]NamedItem, error) {
	rows, err := r.queries.ListDisciplines(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]NamedItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, NamedItem{ID: row.ID, Name: row.Name})
	}
	return result, nil
}

func (r *sqlRepository) ListCompetitionTypes(ctx context.Context) ([]NamedItem, error) {
	rows, err := r.queries.ListCompetitionTypes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]NamedItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, NamedItem{ID: row.ID, Name: row.Name})
	}
	return result, nil
}

func (r *sqlRepository) ListCompetitionLevels(ctx context.Context) ([]CompetitionLevel, error) {
	rows, err := r.queries.ListCompetitionLevels(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CompetitionLevel, 0, len(rows))
	for _, row := range rows {
		result = append(result, CompetitionLevel{
			ID: row.ID, Code: textPointer(row.Code), Name: textPointer(row.Name),
		})
	}
	return result, nil
}

func (r *sqlRepository) ListAgeGroups(ctx context.Context) ([]AgeGroup, error) {
	rows, err := r.queries.ListAgeGroups(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]AgeGroup, 0, len(rows))
	for _, row := range rows {
		result = append(result, AgeGroup{
			ID: row.ID, Name: row.Name, MinAge: int16Pointer(row.MinAge), MaxAge: int16Pointer(row.MaxAge),
		})
	}
	return result, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func int16Pointer(value pgtype.Int2) *int16 {
	if !value.Valid {
		return nil
	}
	return &value.Int16
}
