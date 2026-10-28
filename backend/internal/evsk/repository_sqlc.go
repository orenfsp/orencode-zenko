package evsk

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"time"
)

type sqlRepository struct{ queries *db.Queries }

func NewRepository(q *db.Queries) Repository { return &sqlRepository{queries: q} }
func (r *sqlRepository) Upsert(ctx context.Context, req UpsertRequest) (Standard, error) {
	numeric := pgtype.Numeric{}
	if req.MinimumScore != nil {
		if err := numeric.Scan([]byte(*req.MinimumScore)); err != nil {
			return Standard{}, err
		}
	}
	row, err := r.queries.UpsertEVSKStandard(ctx, db.UpsertEVSKStandardParams{RankID: req.RankID, DisciplineID: req.DisciplineID, MinimumPlace: optionalInt(req.MinimumPlace), MinimumScore: numeric, ValidFrom: optionalDate(req.ValidFrom), ValidTo: optionalDate(req.ValidTo)})
	if err != nil {
		return Standard{}, err
	}
	return fromValues(row.ID, row.RankID, row.DisciplineID, row.MinimumPlace, row.MinimumScore, row.ValidFrom, row.ValidTo, row.CreatedAt), nil
}
func (r *sqlRepository) List(ctx context.Context) ([]Standard, error) {
	rows, err := r.queries.ListEVSKStandards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Standard, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromValues(row.ID, row.RankID, row.DisciplineID, row.MinimumPlace, row.MinimumScore, row.ValidFrom, row.ValidTo, row.CreatedAt))
	}
	return out, nil
}
func fromValues(id, rankID, disciplineID int64, place pgtype.Int4, score string, from, to pgtype.Date, created pgtype.Timestamptz) Standard {
	return Standard{ID: id, RankID: rankID, DisciplineID: disciplineID, MinimumPlace: optionalIntValue(place), MinimumScore: optionalString(score), ValidFrom: optionalDateValue(from), ValidTo: optionalDateValue(to), CreatedAt: created.Time}
}
func optionalInt(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}
func optionalDate(v *string) pgtype.Date {
	if v == nil {
		return pgtype.Date{}
	}
	t, e := time.Parse("2006-01-02", *v)
	if e != nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: t, Valid: true}
}
func optionalIntValue(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}
	x := v.Int32
	return &x
}
func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
func optionalDateValue(v pgtype.Date) *time.Time {
	if !v.Valid {
		return nil
	}
	x := v.Time
	return &x
}
