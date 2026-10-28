package team

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"strings"
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
func (r *sqlRepository) Create(ctx context.Context, name string, members, coaches []int64) (Squad, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Squad{}, err
	}
	defer tx.Rollback(ctx)
	q := r.queries.WithTx(tx)
	row, err := q.CreateTeam(ctx, name)
	if err != nil {
		return Squad{}, err
	}
	for _, id := range members {
		if err = q.AddTeamMember(ctx, db.AddTeamMemberParams{TeamID: row.ID, UserID: id}); err != nil {
			return Squad{}, err
		}
	}
	for _, id := range coaches {
		if err = q.AddTeamCoach(ctx, db.AddTeamCoachParams{TeamID: row.ID, CoachUserID: id}); err != nil {
			return Squad{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Squad{}, err
	}
	return r.load(ctx, row.ID, row.Name, row.CreatedAt), nil
}
func (r *sqlRepository) List(ctx context.Context) ([]Squad, error) {
	rows, err := r.queries.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Squad, 0, len(rows))
	for _, row := range rows {
		result = append(result, r.load(ctx, row.ID, row.Name, row.CreatedAt))
	}
	return result, nil
}
func (r *sqlRepository) load(ctx context.Context, id int64, name string, created pgtype.Timestamptz) Squad {
	s := Squad{ID: id, Name: name, Members: []Member{}, Coaches: []Member{}, CreatedAt: created.Time}
	members, err := r.queries.ListTeamMembers(ctx, id)
	if err == nil {
		for _, m := range members {
			s.Members = append(s.Members, Member{UserID: m.UserID, LastName: m.LastName, FirstName: m.FirstName, MiddleName: textPointer(m.MiddleName)})
		}
	}
	coaches, err := r.queries.ListTeamCoaches(ctx, id)
	if err == nil {
		for _, m := range coaches {
			s.Coaches = append(s.Coaches, Member{UserID: m.CoachUserID, LastName: m.LastName, FirstName: m.FirstName, MiddleName: textPointer(m.MiddleName)})
		}
	}
	return s
}
func textPointer(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := strings.TrimSpace(v.String)
	if s == "" {
		return nil
	}
	return &s
}
