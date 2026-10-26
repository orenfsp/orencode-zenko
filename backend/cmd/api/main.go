package main

import (
	"context"
	"log"
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/config"
	"github.com/orenfsp/orencode-zenko.git/internal/auth"
	"github.com/orenfsp/orencode-zenko.git/pkg/db"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
	"go.uber.org/zap"
)

func main() {
	logger.Init()
	defer func() { _ = logger.Log.Sync() }()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Database.URL == "" {
		logger.Log.Fatal("DATABASE_URL is required")
	}
	if cfg.JWT.Secret == "" {
		logger.Log.Fatal("JWT_SECRET is required")
	}

	pool := db.NewPostgres(context.Background(), cfg.Database.URL)
	defer pool.Close()

	jwtManager := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	repository := auth.NewPostgresRepository(pool)
	service := auth.NewService(repository, jwtManager, repository)
	handler := auth.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", handler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", handler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", handler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", handler.Logout)
	mux.HandleFunc("PATCH /api/v1/profile", handler.PatchProfile)

	address := ":" + cfg.HTTP.Port
	logger.Log.Info("server started", zap.String("address", address))
	if err := http.ListenAndServe(address, mux); err != nil {
		logger.Log.Fatal("http server stopped", zap.Error(err))
	}
}
