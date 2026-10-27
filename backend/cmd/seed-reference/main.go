package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

var officialDisciplines = []string{
	"программирование алгоритмическое",
	"программирование продуктовое",
	"программирование систем информационной безопасности",
	"программирование систем искусственного интеллекта",
	"программирование робототехники",
	"программирование беспилотных авиационных систем",
}

type disciplineUpserter interface {
	UpsertDiscipline(ctx context.Context, name string) error
}

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := seedDisciplines(ctx, db.New(pool).WithTx(tx)); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("reference seed applied; disciplines: %d", len(officialDisciplines))
}

func seedDisciplines(ctx context.Context, upserter disciplineUpserter) error {
	for _, name := range officialDisciplines {
		if err := upserter.UpsertDiscipline(ctx, name); err != nil {
			return err
		}
	}
	return nil
}
