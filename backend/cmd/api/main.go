package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/orenfsp/orencode-zenko.git/config"
	dbsqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/orenfsp/orencode-zenko.git/internal/antidoping"
	"github.com/orenfsp/orencode-zenko.git/internal/audit"
	"github.com/orenfsp/orencode-zenko.git/internal/auth"
	"github.com/orenfsp/orencode-zenko.git/internal/calendarplan"
	"github.com/orenfsp/orencode-zenko.git/internal/competition"
	"github.com/orenfsp/orencode-zenko.git/internal/competitionevent"
	"github.com/orenfsp/orencode-zenko.git/internal/dashboard"
	"github.com/orenfsp/orencode-zenko.git/internal/deadline"
	"github.com/orenfsp/orencode-zenko.git/internal/file"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"github.com/orenfsp/orencode-zenko.git/internal/judgecategory"
	"github.com/orenfsp/orencode-zenko.git/internal/municipality"
	"github.com/orenfsp/orencode-zenko.git/internal/notification" // <-- ДОБАВЛЕНО
	"github.com/orenfsp/orencode-zenko.git/internal/notifier"
	"github.com/orenfsp/orencode-zenko.git/internal/rank"
	rankuser "github.com/orenfsp/orencode-zenko.git/internal/rank_user"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/internal/reference"
	"github.com/orenfsp/orencode-zenko.git/internal/registry"
	"github.com/orenfsp/orencode-zenko.git/internal/result"
	usercoach "github.com/orenfsp/orencode-zenko.git/internal/user_coach"
	postgres "github.com/orenfsp/orencode-zenko.git/pkg/db"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"github.com/orenfsp/orencode-zenko.git/pkg/token"
	"go.uber.org/zap"

	"os"

	"github.com/jackc/pgx/v5"
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

	// --- Запуск Telegram-уведомлений ---
	listenConn, err := pgx.Connect(context.Background(), cfg.Database.URL)
	if err != nil {
		logger.Log.Error("failed to connect for LISTEN: " + err.Error())
	} else {
		go func() {
			tgToken := os.Getenv("TG_TOKEN_BOT")
			if tgToken == "" {
				logger.Log.Warn("TG_TOKEN_BOT is empty, telegram notifier disabled")
				return
			}
			if err := notifier.Start(context.Background(), listenConn, queries, tgToken); err != nil {
				logger.Log.Error("notifier stopped: " + err.Error())
			}
		}()
	}

	jwtManager := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	auditRepository := audit.NewRepository(queries)
	auditService := audit.NewService(auditRepository)
	auditHandler := audit.NewHandler(auditService)
	auditMiddleware := audit.NewMiddleware(auditRepository, jwtManager)
	authRepository := auth.NewRepository(queries, pool)
	authService := auth.NewService(authRepository, jwtManager, authRepository)

	// RBAC.
	rbacRepository := rbac.NewRepository(queries, pool)
	rbacService := rbac.NewService(rbacRepository)
	rbacHandler := rbac.NewHandler(rbacService)

	municipalityRepo := municipality.NewRepository(queries)
	municipalityService := municipality.NewService(municipalityRepo)
	municipalityHandler := municipality.NewHandler(municipalityService)
	rbacMiddleware := rbac.NewMiddleware(jwtManager, rbacRepository)

	protected := func(permission string, handler http.HandlerFunc) http.Handler {
		return rbacMiddleware.Authenticate(rbacMiddleware.Require(permission)(handler))
	}

	// File service.
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

	// Periodic cleanup of pending file deletions.
	cleanupCtx, cancelCleanup := context.WithCancel(context.Background())
	defer cancelCleanup()

	if cfg.Storage.CleanupInterval > 0 {
		go func(ctx context.Context, svc *file.Service, storageCfg *config.StorageConfig) {
			ticker := time.NewTicker(storageCfg.CleanupInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					cutoff := time.Now().Add(-storageCfg.CleanupRetention)
					result, err := svc.CleanupPending(ctx, cutoff, storageCfg.CleanupBatch)
					if err != nil {
						logger.Log.Error(
							"file cleanup failed",
							zap.Error(err),
							zap.Int("processed", result.Processed),
							zap.Int("failed", result.Failed),
						)
						continue
					}
					if result.Processed > 0 || result.Failed > 0 {
						logger.Log.Info(
							"file cleanup finished",
							zap.Int("processed", result.Processed),
							zap.Int("failed", result.Failed),
						)
					}
				}
			}
		}(cleanupCtx, fileService, &cfg.Storage)
	}

	authHandler := auth.NewHandler(authService, fileService)

	fileHandler, err := file.NewHandler(fileService, rbacRepository)
	if err != nil {
		logger.Log.Fatal("initialize file handler", zap.Error(err))
	}

	antiDopingRepository, err := antidoping.NewRepository(queries, pool)
	if err != nil {
		logger.Log.Fatal("initialize anti-doping repository", zap.Error(err))
	}
	antiDopingService := antidoping.NewService(antiDopingRepository)
	antiDopingHandler := antidoping.NewHandler(antiDopingService, fileService)

	antiDopingCtx, cancelAntiDoping := context.WithCancel(context.Background())
	defer cancelAntiDoping()
	go runAntiDopingExpiration(antiDopingCtx, antiDopingService)

	// Competition.
	competitionRepository := competition.NewRepository(queries)
	competitionService := competition.NewService(competitionRepository)
	competitionHandler := competition.NewHandler(competitionService)
	competitionEventRepository := competitionevent.NewRepository(queries, pool)
	competitionEventService := competitionevent.NewService(competitionEventRepository)
	competitionEventHandler := competitionevent.NewHandler(competitionEventService)
	judgeRepository := competition.NewJudgeRepository(queries)
	judgeService := competition.NewJudgeService(judgeRepository)
	judgeHandler := competition.NewJudgeHandler(judgeService)
	resultRepository := result.NewRepository(queries, pool)
	resultService := result.NewService(resultRepository, fileStorage, result.NewPDFProtocolGenerator())
	resultHandler := result.NewHandler(resultService)
	calendarPlanRepository := calendarplan.NewRepository(queries)
	calendarPlanService := calendarplan.NewService(calendarPlanRepository)
	calendarPlanHandler := calendarplan.NewHandler(calendarPlanService)
	dashboardRepository := dashboard.NewRepository(queries)
	dashboardService := dashboard.NewService(dashboardRepository)
	dashboardHandler := dashboard.NewHandler(dashboardService)
	registryRepository := registry.NewRepository(queries)
	registryService := registry.NewService(registryRepository)
	registryHandler := registry.NewHandler(registryService)
	roleChecker := rbac.NewRoleChecker(queries)
	judgeCategoryRepository := judgecategory.NewRepository(queries, pool)
	judgeCategoryService := judgecategory.NewService(judgeCategoryRepository, roleChecker)
	judgeCategoryHandler := judgecategory.NewHandler(judgeCategoryService)
	deadlineRepository := deadline.NewRepository(queries)
	deadlineService := deadline.NewService(deadlineRepository, rbacRepository)
	deadlineHandler := deadline.NewHandler(deadlineService)

	// Reference data.
	referenceRepository := reference.NewRepository(queries)
	referenceService := reference.NewService(referenceRepository)
	referenceHandler := reference.NewHandler(referenceService)

	// User-coach relations.
	userCoachRepo := usercoach.NewSQLRepository(queries)

	userCoachService, err := usercoach.NewService(userCoachRepo, roleChecker)
	if err != nil {
		logger.Log.Fatal("initialize user coach service", zap.Error(err))
	}
	userCoachHandler := usercoach.NewHandler(userCoachService)

	// Rankings.
	rankRepository := rank.NewRepository(queries)
	rankService := rank.NewService(rankRepository)
	rankHandler := rank.NewHandler(rankService)
	cleanupExpiredRanks := func(ctx context.Context) {
		deleted, err := rankService.CleanupExpired(ctx)
		if err != nil {
			logger.Log.Error("rank cleanup failed", zap.Error(err))
			return
		}
		if deleted > 0 {
			logger.Log.Info("expired ranks deleted", zap.Int64("count", deleted))
		}
	}
	cleanupExpiredRanks(cleanupCtx)
	if cfg.Rank.CleanupInterval > 0 {
		go func(ctx context.Context, interval time.Duration) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					cleanupExpiredRanks(ctx)
				}
			}
		}(cleanupCtx, cfg.Rank.CleanupInterval)
	}
	rankApplicationRepository := rankuser.NewRepository(queries)
	rankApplicationService := rankuser.NewService(rankApplicationRepository)
	rankApplicationHandler := rankuser.NewHandler(rankApplicationService)

	// Adds the authenticated user's ID to the context expected by usercoach.Handler.
	// The token is verified by the existing RBAC middleware first.
	withUserCoachIdentity := func(next http.Handler) http.Handler {
		return rbacMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const bearerPrefix = "Bearer "
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))

			if len(authorization) <= len(bearerPrefix) ||
				!strings.EqualFold(authorization[:len(bearerPrefix)], bearerPrefix) {
				http.Error(w, "требуется авторизация", http.StatusUnauthorized)
				return
			}

			rawToken := strings.TrimSpace(authorization[len(bearerPrefix):])
			if rawToken == "" || strings.ContainsAny(rawToken, " \t\r\n") {
				http.Error(w, "недействительный токен", http.StatusUnauthorized)
				return
			}

			claims, err := jwtManager.Parse(rawToken)
			if err != nil || claims == nil ||
				claims.Type != "access" ||
				claims.PrincipalType != "user" ||
				claims.UserID <= 0 {
				http.Error(w, "недействительный токен", http.StatusUnauthorized)
				return
			}

			ctx := usercoach.WithAuthenticatedUser(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}

	// --- Инициализация уведомлений для фронта ---
	notificationRepo := notification.NewRepository(queries)
	notificationHandler := notification.NewHandler(notificationRepo)

	// HTTP routes.
	mux := http.NewServeMux()

	// Authentication.
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("PATCH /api/v1/profile", authHandler.PatchProfile)

	// Files.
	mux.Handle("GET /api/v1/files/{fileID}/download", rbacMiddleware.Authenticate(http.HandlerFunc(fileHandler.Download)))
	mux.Handle("GET /api/v1/admin/files", protected(rbac.PermissionFilesRead, fileHandler.ListAll))
	registerAntiDopingRoutes(
		mux,
		antiDopingHandler,
		func(handler http.HandlerFunc) http.Handler {
			return rbacMiddleware.Authenticate(handler)
		},
		protected,
	)

	// RBAC administration.
	mux.Handle("GET /api/v1/admin/roles", protected(rbac.PermissionRolesRead, rbacHandler.ListRoles))
	mux.Handle("POST /api/v1/admin/roles", protected(rbac.PermissionRolesCreate, rbacHandler.CreateRole))
	mux.Handle("POST /api/v1/admin/users/{userID}/roles", protected(rbac.PermissionRolesAssign, rbacHandler.AssignRole))
	mux.Handle("GET /api/v1/admin/permissions", protected(rbac.PermissionPermissionsRead, rbacHandler.ListPermissions))
	mux.Handle("GET /api/v1/roles/self-assignable", rbacMiddleware.Authenticate(http.HandlerFunc(rbacHandler.ListSelfAssignableRoles)))

	// User-coach relations.
	mux.Handle("POST /api/v1/coach/user-coaches", withUserCoachIdentity(http.HandlerFunc(userCoachHandler.Create)))
	mux.Handle("GET /api/v1/coach/user-coaches", withUserCoachIdentity(http.HandlerFunc(userCoachHandler.ListByCoach)))
	mux.Handle("GET /api/v1/participant/coaches", withUserCoachIdentity(http.HandlerFunc(userCoachHandler.ListByParticipant)))
	mux.Handle("PATCH /api/v1/user-coaches/{id}/confirm", withUserCoachIdentity(http.HandlerFunc(userCoachHandler.Confirm)))
	mux.Handle("PATCH /api/v1/user-coaches/{id}/end", withUserCoachIdentity(http.HandlerFunc(userCoachHandler.End)))
	mux.Handle("GET /api/v1/admin/audit-logs", protected(rbac.PermissionRolesRead, auditHandler.List))
	// Competition, calendar plans and reference data.
	registerCompetitionRoutes(mux, competitionHandler, protected)
	registerCompetitionEventRoutes(mux, competitionEventHandler, protected)
	registerCompetitionJudgeRoutes(mux, judgeHandler, protected)
	registerCompetitionResultRoutes(mux, resultHandler, protected)
	mux.Handle("GET /api/v1/profile/competition-results", rbacMiddleware.Authenticate(http.HandlerFunc(resultHandler.History)))
	registerCalendarPlanRoutes(mux, calendarPlanHandler, protected)
	registerReferenceRoutes(mux, referenceHandler)
	registerDashboardRoutes(mux, dashboardHandler, protected)
	registerRegistryRoutes(mux, registryHandler, protected)
	registerJudgeCategoryRoutes(
		mux,
		judgeCategoryHandler,
		func(handler http.HandlerFunc) http.Handler { return rbacMiddleware.Authenticate(handler) },
		protected,
	)
	registerDeadlineRoutes(
		mux,
		deadlineHandler,
		func(handler http.HandlerFunc) http.Handler { return rbacMiddleware.Authenticate(handler) },
		protected,
	)

	// Rankings.
	mux.HandleFunc("POST /api/v1/admin/rank", rankHandler.CreateRank)
	mux.HandleFunc("GET /api/v1/rank/{id}", rankHandler.GetRank)
	mux.HandleFunc("GET /api/v1/rank/", rankHandler.ListRank)
	registerRankApplicationRoutes(
		mux,
		rankApplicationHandler,
		func(handler http.HandlerFunc) http.Handler { return rbacMiddleware.Authenticate(handler) },
		protected,
	)

	mux.HandleFunc("GET /api/v1/municipalities", municipalityHandler.List)
	mux.HandleFunc("GET /api/v1/municipalities/{id}", municipalityHandler.Get)
	mux.Handle("POST /api/v1/admin/municipalities",
		protected(rbac.PermissionMunicipalitiesCreate, municipalityHandler.Create))

	// --- Роуты для уведомлений ---
	// Используем rbacMiddleware.Authenticate, чтобы проверить токен (когда друг подключит JWT)
	mux.Handle("GET /api/v1/notifications", rbacMiddleware.Authenticate(http.HandlerFunc(notificationHandler.ListNotifications)))
	mux.Handle("PATCH /api/v1/notifications/{id}/read", rbacMiddleware.Authenticate(http.HandlerFunc(notificationHandler.MarkAsRead)))
	mux.Handle("GET /api/v1/profile/competitions",
		rbacMiddleware.Authenticate(http.HandlerFunc(judgeHandler.ListByUser)))
	mux.Handle("GET /api/v1/profile/assignments",
		rbacMiddleware.Authenticate(http.HandlerFunc(judgeHandler.ListAssignments)))

	address := ":" + cfg.HTTP.Port
	handler := httpx.CORS(cfg.HTTP.CORSAllowedOrigins)(auditMiddleware.Wrap(mux))

	logger.Log.Info("server started", zap.String("address", address))
	if err := http.ListenAndServe(address, handler); err != nil {
		logger.Log.Fatal("http server stopped", zap.Error(err))
	}
}

func runAntiDopingExpiration(ctx context.Context, service *antidoping.Service) {
	expire := func() {
		rows, err := service.Expire(ctx)
		if err != nil {
			if ctx.Err() == nil {
				logger.Log.Error("anti-doping expiration failed", zap.Error(err))
			}
			return
		}
		if rows > 0 {
			logger.Log.Info("anti-doping certificates expired", zap.Int64("count", rows))
		}
	}

	expire()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expire()
		}
	}
}
