package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/orenfsp/orencode-zenko.git/config"
	dbsqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/orenfsp/orencode-zenko.git/internal/auth"
	"github.com/orenfsp/orencode-zenko.git/internal/file"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	postgres "github.com/orenfsp/orencode-zenko.git/pkg/db"
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

	pool := postgres.NewPostgres(context.Background(), cfg.Database.URL)
	defer pool.Close()
	queries := dbsqlc.New(pool)

	jwtManager := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	authRepository := auth.NewRepository(queries, pool)
	service := auth.NewService(authRepository, jwtManager, authRepository)
	rbacRepository := rbac.NewRepository(queries, pool)
	rbacService := rbac.NewService(rbacRepository)
	rbacHandler := rbac.NewHandler(rbacService)
	rbacMiddleware := rbac.NewMiddleware(jwtManager, rbacRepository)
	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return rbacMiddleware.Authenticate(rbacMiddleware.Require(permission)(handler))
	}
	fileRepo, err := file.NewSQLCRepository(queries)
	if err != nil {
		logger.Log.Fatal("initialize file repository", zap.Error(err))
	}
	fileStorage, err := file.NewLocalStorage(cfg.Storage.Path)
	if err != nil {
		logger.Log.Fatal("initialize file storage", zap.Error(err))
	}
	fileService, err := file.NewService(fileRepo, fileStorage, file.DefaultPolicies())
	if err != nil {
		logger.Log.Fatal("initialize file service", zap.Error(err))
	}

	cleanupCtx, cancelCleanup := context.WithCancel(context.Background())
	defer cancelCleanup()

	go func(ctx context.Context, svc *file.Service, cfg config.StorageConfig) {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-cfg.CleanupRetention)
				result, err := svc.CleanupPending(ctx, cutoff, cfg.CleanupBatch)
				if err != nil {
					logger.Log.Error("file cleanup failed", zap.Error(err),
						zap.Int("processed", result.Processed), zap.Int("failed", result.Failed))
					continue
				}
				if result.Processed > 0 {
					logger.Log.Info("file cleanup finished",
						zap.Int("processed", result.Processed), zap.Int("failed", result.Failed))
				}
			}
		}
	}(cleanupCtx, fileService, cfg.Storage)

	authHandler := auth.NewHandler(service, fileService)
	fileHandler, err := file.NewHandler(fileService, rbacRepository)
	if err != nil {
		logger.Log.Fatal("initialize file handler", zap.Error(err))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("PATCH /api/v1/profile", authHandler.PatchProfile)
	mux.Handle("GET /api/v1/files/{fileID}/download", rbacMiddleware.Authenticate(http.HandlerFunc(fileHandler.Download)))
	mux.Handle("GET /api/v1/admin/roles", protected(rbac.PermissionRolesRead, rbacHandler.ListRoles))
	mux.Handle("POST /api/v1/admin/roles", protected(rbac.PermissionRolesCreate, rbacHandler.CreateRole))
	mux.Handle("POST /api/v1/admin/users/{userID}/roles", protected(rbac.PermissionRolesAssign, rbacHandler.AssignRole))
	mux.Handle("GET /api/v1/admin/permissions", protected(rbac.PermissionPermissionsRead, rbacHandler.ListPermissions))

	address := ":" + cfg.HTTP.Port
	logger.Log.Info("server started", zap.String("address", address))
	if err := http.ListenAndServe(address, mux); err != nil {
		logger.Log.Fatal("http server stopped", zap.Error(err))
	}

	handler := httpx.CORS(cfg.HTTP.CORSAllowedOrigins)(mux)
	if err := http.ListenAndServe(address, handler); err != nil {
		logger.Log.Fatal("http server stopped", zap.Error(err))
	}
}
