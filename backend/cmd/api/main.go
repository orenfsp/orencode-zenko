package main

import (
	"context"
	"log"
	"github.com/orenfsp/orencode-zenko.git/config"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"github.com/orenfsp/orencode-zenko.git/pkg/db"
sqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	//"github.com/orenfsp/orencode-zenko.git/pkg/token"
	// "github.com/orenfsp/orencode-zenko.git/pkg/audit"
	// "github.com/orenfsp/orencode-zenko.git/pkg/auth"
)

func main() {

	ctx := context.Background()
	//подключение бд
	logger.Init()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	//token manager
	// jwtManager := token.New(
	// 	cfg.JWT.Secret,
	// 	cfg.JWT.AccessTTL,
	// 	cfg.JWT.RefreshTTL,
	// )

	dsn := cfg.Database.URL
	if dsn == "" {
		logger.Log.Fatal("not connetion db, pls chek dsn")
	} else {
		logger.Log.Info("db connection succes")
	}
	pool := db.NewPostgres(ctx, dsn)
	defer pool.Close()
	queries := sqlc.New(pool)

	defer logger.Log.Sync()
	logger.Log.Info("server started on port " + cfg.HTTP.Port)
	_ = queries

	// repoAudit := audit.NewRepository(queries)
	// serviceAudit := audit.NewService(repoAudit)
	// handlerAudit := audit.NewHandler(serviceAudit)

	// authService := auth.NewService(
	// 	repoUser,
	// 	jwtManager,
	// 	repoSession,
	// )
	// authHandler := auth.NewHandler(authService)
}