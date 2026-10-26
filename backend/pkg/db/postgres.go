package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context, dsn string) *pgxpool.Pool {
	pool,err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("failed connestion db check env mb connection string:", err)
	}
	return pool
}
