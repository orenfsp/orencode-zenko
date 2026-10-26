package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/orenfsp/orencode-zenko.git/config"
	sqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/orenfsp/orencode-zenko.git/internal/notifier"
	"github.com/orenfsp/orencode-zenko.git/pkg/db"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
)

func main() {
	ctx := context.Background()

	// Инициализация логгера и конфига
	logger.Init()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Подключение к БД (пул)
	dsn := cfg.Database.URL
	if dsn == "" {
		logger.Log.Fatal("not connetion db, pls chek dsn")
	} else {
		logger.Log.Info("db connection succes")
	}
	pool := db.NewPostgres(ctx, dsn)
	defer pool.Close()

	// Инициализация sqlc
	queries := sqlc.New(pool)
	defer logger.Log.Sync()

	logger.Log.Info("server started on port " + cfg.HTTP.Port)

	// --- Запуск Telegram-уведомлений ---
	// 1. Создаем отдельное соединение для LISTEN (оно не может быть из пула)
	listenConn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect for LISTEN: %v", err)
	}

	// 2. Запускаем notifier в горутине
	go func() {
		tgToken := os.Getenv("TG_TOKEN")
		if tgToken == "" {
			logger.Log.Warn("TG_TOKEN is empty, telegram notifier disabled")
			return
		}
		// Передаем queries (а не отдельное соединение), так как запросы теперь через sqlc
		if err := notifier.Start(ctx, listenConn, queries, tgToken); err != nil {
			logger.Log.Error("notifier stopped: " + err.Error())
		}
	}()

	// 3. Блокируем main, чтобы сервер не завершался с кодом 0
	// (Пока друг не добавил HTTP-сервер, оставляем select{})
	select {}
}