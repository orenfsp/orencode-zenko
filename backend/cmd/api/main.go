package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/orenfsp/orencode-zenko.git/config"
	dbsqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/orenfsp/orencode-zenko.git/internal/antidoping"
	"github.com/orenfsp/orencode-zenko.git/internal/audit"
	"github.com/orenfsp/orencode-zenko.git/internal/auth"
	"github.com/orenfsp/orencode-zenko.git/internal/calendarplan"
	"github.com/orenfsp/orencode-zenko.git/internal/competition"
	"github.com/orenfsp/orencode-zenko.git/internal/competitionapplication"
	"github.com/orenfsp/orencode-zenko.git/internal/competitionevent"
	"github.com/orenfsp/orencode-zenko.git/internal/dashboard"
	"github.com/orenfsp/orencode-zenko.git/internal/deadline"
	"github.com/orenfsp/orencode-zenko.git/internal/evsk"
	"github.com/orenfsp/orencode-zenko.git/internal/file"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"github.com/orenfsp/orencode-zenko.git/internal/judgecategory"
	"github.com/orenfsp/orencode-zenko.git/internal/municipality"
	"github.com/orenfsp/orencode-zenko.git/internal/notification" // <-- ДОБАВЛЕНО
	"github.com/orenfsp/orencode-zenko.git/internal/notifier"
	"github.com/orenfsp/orencode-zenko.git/internal/rank"
	rankuser "github.com/orenfsp/orencode-zenko.git/internal/rank_user"
	"github.com/orenfsp/orencode-zenko.git/internal/rating"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/internal/reference"
	"github.com/orenfsp/orencode-zenko.git/internal/registry"
	"github.com/orenfsp/orencode-zenko.git/internal/result"
	"github.com/orenfsp/orencode-zenko.git/internal/team"
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
	authenticated := func(handler http.HandlerFunc) http.Handler {
		return rbacMiddleware.Authenticate(http.HandlerFunc(handler))
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
	competitionApplicationRepository := competitionapplication.NewRepository(queries)
	competitionApplicationService := competitionapplication.NewService(competitionApplicationRepository)
	competitionApplicationHandler := competitionapplication.NewHandler(competitionApplicationService)
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
	registryReportHandler := registry.NewReportHandler(registryService, fileService)
	ratingRepository := rating.NewRepository(queries)
	ratingService := rating.NewService(ratingRepository)
	ratingHandler := rating.NewHandler(ratingService)
	roleChecker := rbac.NewRoleChecker(queries)
	judgeScoped := func(permission string, handler http.HandlerFunc) http.Handler {
		return rbacMiddleware.Authenticate(rbacMiddleware.Require(permission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := rbac.IdentityFromContext(r.Context())
			if !ok || identity.UserID <= 0 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			isAdmin, err := roleChecker.HasRole(r.Context(), identity.UserID, "admin")
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if !isAdmin {
				isAdmin, err = roleChecker.HasRole(r.Context(), identity.UserID, "super_admin")
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			}
			if !isAdmin {
				isJudge, err := roleChecker.HasRole(r.Context(), identity.UserID, "judge")
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				competitionID, parseErr := strconv.ParseInt(r.PathValue("competitionID"), 10, 64)
				if !isJudge || parseErr != nil || competitionID <= 0 {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				assignments, err := judgeRepository.ListAdminCompetitionJudges(r.Context(), competitionID)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				assigned := false
				for _, assignment := range assignments {
					if assignment.JudgeUserID == identity.UserID {
						assigned = true
						break
					}
				}
				if !assigned {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}
			handler(w, r)
		})))
	}
	teamRepository := team.NewRepository(queries, pool)
	judgeCategoryRepository := judgecategory.NewRepository(queries, pool)
	judgeCategoryService := judgecategory.NewService(judgeCategoryRepository, roleChecker)
	judgeCategoryHandler := judgecategory.NewHandler(judgeCategoryService)
	deadlineRepository := deadline.NewRepository(queries)
	deadlineService := deadline.NewService(deadlineRepository, rbacRepository)
	deadlineHandler := deadline.NewHandler(deadlineService)
	deadlineCtx, cancelDeadline := context.WithCancel(context.Background())
	defer cancelDeadline()
	go deadline.RunReminderWorker(deadlineCtx, deadlineRepository, cfg.Deadline.ReminderInterval, cfg.Deadline.ReminderDays)

	// Reference data.
	referenceRepository := reference.NewRepository(queries)
	referenceService := reference.NewService(referenceRepository)
	referenceHandler := reference.NewHandler(referenceService)

	// User-coach relations.
	userCoachRepo := usercoach.NewSQLRepository(queries)
	teamService := team.NewService(teamRepository, roleChecker, userCoachRepo)
	teamHandler := team.NewHandler(teamService)

	userCoachService, err := usercoach.NewService(userCoachRepo, roleChecker)
	if err != nil {
		logger.Log.Fatal("initialize user coach service", zap.Error(err))
	}
	userCoachHandler := usercoach.NewHandler(userCoachService)

	// Rankings.
	rankRepository := rank.NewRepository(queries)
	rankService := rank.NewService(rankRepository)
	rankHandler := rank.NewHandler(rankService)
	evskService := evsk.NewService(evsk.NewRepository(queries))
	evskHandler := evsk.NewHandler(evskService)
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
	mux.Handle("GET /api/v1/profile", authenticated(authHandler.GetProfile))
	mux.HandleFunc("GET /api/v1/profiles/{userID}", authHandler.GetPublicProfile)
	mux.HandleFunc("PATCH /api/v1/profile", authHandler.PatchProfile)

	// Files.
	mux.Handle("GET /api/v1/files/{fileID}/download", rbacMiddleware.Authenticate(http.HandlerFunc(fileHandler.Download)))
	mux.Handle("GET /api/v1/admin/files", protected(rbac.PermissionFilesRead, fileHandler.ListAll))
	mux.Handle("POST /api/v1/anti-doping/self", authenticated(antiDopingHandler.Upload))
	mux.Handle("GET /api/v1/anti-doping/self", authenticated(antiDopingHandler.ListMine))
	mux.Handle("GET /api/v1/admin/anti-doping", protected(rbac.PermissionAntiDopingReview, antiDopingHandler.ListAll))
	mux.Handle("PATCH /api/v1/admin/anti-doping/{certificateID}", protected(rbac.PermissionAntiDopingReview, antiDopingHandler.Review))

	// RBAC administration.
	mux.Handle("GET /api/v1/admin/roles", protected(rbac.PermissionRolesRead, rbacHandler.ListRoles))
	mux.Handle("POST /api/v1/admin/roles", protected(rbac.PermissionRolesCreate, rbacHandler.CreateRole))
	mux.Handle("POST /api/v1/admin/users/{userID}/roles", protected(rbac.PermissionRolesAssign, rbacHandler.AssignRole))
	mux.Handle("GET /api/v1/admin/permissions", protected(rbac.PermissionPermissionsRead, rbacHandler.ListPermissions))
	mux.Handle("GET /api/v1/my-permissions", authenticated(rbacHandler.MyPermissions))
	mux.Handle("GET /api/v1/my-permission", authenticated(rbacHandler.MyPermissions))
	mux.Handle("GET /api/v1/roles/self-assignable", rbacMiddleware.Authenticate(http.HandlerFunc(rbacHandler.ListSelfAssignableRoles)))

	// User-coach relations.
	mux.Handle("POST /api/v1/coach/user-coaches", rbacMiddleware.Authenticate(http.HandlerFunc(userCoachHandler.Create)))
	mux.Handle("GET /api/v1/coach/user-coaches", rbacMiddleware.Authenticate(http.HandlerFunc(userCoachHandler.ListByCoach)))
	mux.Handle("GET /api/v1/participant/coaches", rbacMiddleware.Authenticate(http.HandlerFunc(userCoachHandler.ListByParticipant)))
	mux.Handle("PATCH /api/v1/user-coaches/{id}/confirm", rbacMiddleware.Authenticate(http.HandlerFunc(userCoachHandler.Confirm)))
	mux.Handle("PATCH /api/v1/user-coaches/{id}/end", rbacMiddleware.Authenticate(http.HandlerFunc(userCoachHandler.End)))
	mux.Handle("GET /api/v1/admin/audit-logs", protected(rbac.PermissionRolesRead, auditHandler.List))
	// Competition, calendar plans and reference data.
	mux.HandleFunc("GET /api/v1/calendar/competitions", competitionHandler.Calendar)
	mux.HandleFunc("GET /api/v1/competitions", competitionHandler.List)
	mux.HandleFunc("GET /api/v1/competitions/{id}", competitionHandler.Get)
	mux.Handle("POST /api/v1/competitions", protected(competition.PermissionCreate, competitionHandler.Create))
	mux.Handle("PATCH /api/v1/competitions/{id}", protected(competition.PermissionUpdate, competitionHandler.Update))
	mux.Handle("GET /api/v1/admin/competitions", protected(competition.PermissionUpdate, competitionHandler.ListAdmin))
	mux.Handle("GET /api/v1/admin/competitions/{id}", protected(competition.PermissionUpdate, competitionHandler.GetAdmin))
	mux.Handle("POST /api/v1/admin/competitions/{id}/submit", protected(competition.PermissionUpdate, competitionHandler.Submit))
	mux.Handle("POST /api/v1/admin/competitions/{id}/publish", protected(competition.PermissionUpdate, competitionHandler.Publish))
	mux.Handle("POST /api/v1/admin/competitions/{id}/cancel", protected(competition.PermissionUpdate, competitionHandler.Cancel))
	mux.Handle("POST /api/v1/admin/competitions/{id}/return", protected(competition.PermissionUpdate, competitionHandler.ReturnToDraft))
	mux.Handle("GET /api/v1/admin/competitions/{id}/history", protected(competition.PermissionUpdate, competitionHandler.History))
	mux.Handle("POST /api/v1/competition-applications", authenticated(competitionApplicationHandler.Create))
	mux.Handle("GET /api/v1/competition-applications/my", authenticated(competitionApplicationHandler.ListMine))
	mux.Handle("GET /api/v1/competition-applications/{id}", authenticated(competitionApplicationHandler.GetMine))
	mux.Handle("POST /api/v1/competition-applications/{id}/withdraw", authenticated(competitionApplicationHandler.Withdraw))
	mux.Handle("GET /api/v1/admin/competition-applications", protected(competitionapplication.PermissionReview, competitionApplicationHandler.ListForReview))
	mux.Handle("GET /api/v1/admin/competition-applications/{id}", protected(competitionapplication.PermissionReview, competitionApplicationHandler.GetForReview))
	mux.Handle("POST /api/v1/admin/competition-applications/{id}/approve", protected(competitionapplication.PermissionReview, competitionApplicationHandler.Approve))
	mux.Handle("POST /api/v1/admin/competition-applications/{id}/reject", protected(competitionapplication.PermissionReview, competitionApplicationHandler.Reject))
	mux.HandleFunc("GET /api/v1/competitions/{competitionID}/events", competitionEventHandler.ListPublic)
	mux.HandleFunc("GET /api/v1/competitions/{competitionID}/events/{eventID}", competitionEventHandler.GetPublic)
	mux.Handle("GET /api/v1/admin/competitions/{competitionID}/events", protected(competitionevent.PermissionRead, competitionEventHandler.ListAdmin))
	mux.Handle("GET /api/v1/admin/competitions/{competitionID}/events/{eventID}", protected(competitionevent.PermissionRead, competitionEventHandler.GetAdmin))
	mux.Handle("POST /api/v1/competitions/{competitionID}/events", protected(competitionevent.PermissionCreate, competitionEventHandler.Create))
	mux.Handle("PATCH /api/v1/competitions/{competitionID}/events/{eventID}", protected(competitionevent.PermissionUpdate, competitionEventHandler.Update))
	mux.Handle("DELETE /api/v1/competitions/{competitionID}/events/{eventID}", protected(competitionevent.PermissionDelete, competitionEventHandler.Delete))
	mux.HandleFunc("GET /api/v1/competitions/{id}/judges", judgeHandler.List)
	mux.Handle("GET /api/v1/admin/competitions/{id}/judges", protected(competition.PermissionManageJudges, judgeHandler.ListAdmin))
	mux.Handle("POST /api/v1/competitions/{id}/judges", protected(competition.PermissionManageJudges, judgeHandler.Assign))
	mux.Handle("DELETE /api/v1/competitions/{id}/judges/{assignmentId}", protected(competition.PermissionManageJudges, judgeHandler.Remove))
	resultBase := "/api/v1/admin/competitions/{competitionID}/events/{eventID}/results"
	mux.Handle("GET "+resultBase, judgeScoped(result.PermissionRead, resultHandler.Get))
	mux.Handle("PUT "+resultBase+"/settings", judgeScoped(result.PermissionManage, resultHandler.Configure))
	mux.Handle("PUT "+resultBase, judgeScoped(result.PermissionManage, resultHandler.Save))
	mux.Handle("POST "+resultBase+"/calculate", judgeScoped(result.PermissionManage, resultHandler.Calculate))
	mux.Handle("POST "+resultBase+"/approve", judgeScoped(result.PermissionApprove, resultHandler.Approve))
	mux.HandleFunc("GET /api/v1/competitions/{competitionID}/events/{eventID}/protocol", resultHandler.DownloadProtocol)
	mux.Handle("GET /api/v1/profile/competition-results", rbacMiddleware.Authenticate(http.HandlerFunc(resultHandler.History)))
	mux.Handle("GET /api/v1/admin/calendar-plans", protected(calendarplan.PermissionRead, calendarPlanHandler.List))
	mux.Handle("GET /api/v1/admin/calendar-plans/{id}", protected(calendarplan.PermissionRead, calendarPlanHandler.Get))
	mux.Handle("POST /api/v1/admin/calendar-plans", protected(calendarplan.PermissionCreate, calendarPlanHandler.Create))
	mux.Handle("PATCH /api/v1/admin/calendar-plans/{id}", protected(calendarplan.PermissionUpdate, calendarPlanHandler.Update))
	mux.Handle("GET /api/v1/admin/calendar-plans/{id}/competitions", protected(calendarplan.PermissionRead, calendarPlanHandler.ListCompetitions))
	mux.HandleFunc("GET /api/v1/disciplines", referenceHandler.ListDisciplines)
	mux.HandleFunc("GET /api/v1/competition-types", referenceHandler.ListCompetitionTypes)
	mux.HandleFunc("GET /api/v1/competition-levels", referenceHandler.ListCompetitionLevels)
	mux.HandleFunc("GET /api/v1/age-groups", referenceHandler.ListAgeGroups)
	mux.Handle("GET /api/v1/admin/dashboard", protected(dashboard.PermissionRead, dashboardHandler.Get))
	mux.Handle("GET /api/v1/admin/registry/users", protected(registry.PermissionRead, registryHandler.ListUsers))
	mux.Handle("GET /api/v1/admin/registry/organizations", protected(registry.PermissionRead, registryHandler.ListOrganizations))
	mux.Handle("GET /api/v1/admin/reports/users.xlsx", protected(registry.PermissionRead, registryReportHandler.UsersXLSX))
	mux.Handle("GET /api/v1/admin/rating", protected(rating.PermissionRead, ratingHandler.List))
	mux.Handle("GET /api/v1/admin/teams", protected(team.PermissionManage, teamHandler.List))
	mux.Handle("POST /api/v1/admin/teams", protected(team.PermissionManage, teamHandler.Create))
	mux.Handle("POST /api/v1/judge-category-applications", authenticated(judgeCategoryHandler.Create))
	mux.Handle("GET /api/v1/judge-category-applications/mine", authenticated(judgeCategoryHandler.ListMine))
	mux.Handle("GET /api/v1/judge-category/current", authenticated(judgeCategoryHandler.GetCurrent))
	mux.Handle("GET /api/v1/judge/competitions", authenticated(judgeCategoryHandler.ListAssignments))
	mux.Handle("GET /api/v1/admin/judge-category-applications", protected(judgecategory.PermissionReview, judgeCategoryHandler.ListPending))
	mux.Handle("PATCH /api/v1/admin/judge-category-applications/{id}", protected(judgecategory.PermissionReview, judgeCategoryHandler.Review))
	mux.Handle("GET /api/v1/deadlines", authenticated(deadlineHandler.List))
	mux.Handle("GET /api/v1/admin/report-deadlines", protected(deadline.PermissionManage, deadlineHandler.ListReports))
	mux.Handle("POST /api/v1/admin/report-deadlines", protected(deadline.PermissionManage, deadlineHandler.CreateReport))
	mux.Handle("PATCH /api/v1/admin/report-deadlines/{id}", protected(deadline.PermissionManage, deadlineHandler.UpdateReport))

	// Rankings.
	mux.Handle("POST /api/v1/admin/rank", protected(rbac.PermissionRankApplicationsReview, rankHandler.CreateRank))
	mux.HandleFunc("GET /api/v1/rank/{id}", rankHandler.GetRank)
	mux.HandleFunc("GET /api/v1/rank", rankHandler.ListRank)
	mux.Handle("POST /api/v1/rank-applications", authenticated(rankApplicationHandler.Create))
	mux.Handle("GET /api/v1/rank-applications/mine", authenticated(rankApplicationHandler.ListMine))
	mux.Handle("GET /api/v1/ranks/history", authenticated(rankApplicationHandler.ListHistory))
	mux.Handle("GET /api/v1/admin/rank-applications", protected(rbac.PermissionRankApplicationsReview, rankApplicationHandler.ListPending))
	mux.Handle("PATCH /api/v1/admin/rank-applications/{applicationID}", protected(rbac.PermissionRankApplicationsReview, rankApplicationHandler.Review))
	mux.Handle("GET /api/v1/admin/rank-standards", protected(rbac.PermissionRankApplicationsReview, evskHandler.List))
	mux.Handle("POST /api/v1/admin/rank-standards", protected(rbac.PermissionRankApplicationsReview, evskHandler.Upsert))

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
