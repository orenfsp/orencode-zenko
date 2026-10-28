package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/password"
)

type adminSeedConfig struct {
	DatabaseURL string
	Email       string
	Password    string
	FirstName   string
	LastName    string
	BirthDate   time.Time
}

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	if err := seed(ctx, pool, config); err != nil {
		log.Fatal(err)
	}
	log.Printf("RBAC seed applied; super administrator: %s", config.Email)
}

func loadConfig() (adminSeedConfig, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return adminSeedConfig{}, errors.New("DATABASE_URL is required")
	}

	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	if email == "" {
		legacyName := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_NAME")))
		if legacyName == "" {
			legacyName = "admin"
		}
		if strings.Contains(legacyName, "@") {
			email = legacyName
		} else {
			email = legacyName + "@local.test"
		}
	}
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return adminSeedConfig{}, errors.New("ADMIN_EMAIL must be a valid email address")
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if len(adminPassword) < 12 || len(adminPassword) > 72 {
		return adminSeedConfig{}, errors.New("ADMIN_PASSWORD must contain from 12 to 72 bytes")
	}

	birthDateValue := envOrDefault("ADMIN_BIRTH_DATE", "1970-01-01")
	birthDate, err := time.Parse("2006-01-02", birthDateValue)
	if err != nil {
		return adminSeedConfig{}, fmt.Errorf("ADMIN_BIRTH_DATE: %w", err)
	}

	return adminSeedConfig{
		DatabaseURL: databaseURL,
		Email:       email,
		Password:    adminPassword,
		FirstName:   envOrDefault("ADMIN_FIRST_NAME", "Главный"),
		LastName:    envOrDefault("ADMIN_LAST_NAME", "Администратор"),
		BirthDate:   birthDate,
	}, nil
}

func seed(ctx context.Context, pool *pgxpool.Pool, config adminSeedConfig) error {
	passwordHash, err := password.Hash(config.Password)
	if err != nil {
		return err
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := db.New(pool).WithTx(tx)

	permissionIDs := make([]int64, 0, len(rbac.DefaultPermissions))
	for _, permission := range rbac.DefaultPermissions {
		permissionID, err := queries.UpsertPermission(ctx, db.UpsertPermissionParams{
			Code:        permission.Code,
			Description: permission.Description,
		})
		if err != nil {
			return err
		}
		permissionIDs = append(permissionIDs, permissionID)
	}

	roleID, err := queries.UpsertSuperAdminRole(ctx, db.UpsertSuperAdminRoleParams{
		Code: rbac.SuperAdminRoleCode,
		Name: "Главный администратор",
	})
	if err != nil {
		return err
	}
	adminRoleID, err := queries.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{
		Code: rbac.AdminRoleCode,
		Name: "Администратор",
	})
	if err != nil {
		return err
	}
	for _, role := range rbac.DefaultSelfAssignableRoles {
		if _, err := queries.UpsertSelfAssignableRole(ctx, db.UpsertSelfAssignableRoleParams{
			Code: role.Code,
			Name: role.Name,
		}); err != nil {
			return err
		}
	}

	for _, permissionID := range permissionIDs {
		if err := queries.AssignRolePermission(ctx, db.AssignRolePermissionParams{
			RoleID:       roleID,
			PermissionID: permissionID,
		}); err != nil {
			return err
		}
		if err := queries.AssignRolePermission(ctx, db.AssignRolePermissionParams{
			RoleID:       adminRoleID,
			PermissionID: permissionID,
		}); err != nil {
			return err
		}
	}

	userID, err := queries.UpsertSeedAdminUser(ctx, db.UpsertSeedAdminUserParams{
		FirstName:    config.FirstName,
		LastName:     config.LastName,
		Birth:        pgtype.Date{Time: config.BirthDate, Valid: true},
		Email:        config.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return err
	}

	if err := queries.AssignRole(ctx, db.AssignRoleParams{UserID: userID, RoleID: roleID}); err != nil {
		return err
	}
	if err := queries.AssignRole(ctx, db.AssignRoleParams{UserID: userID, RoleID: adminRoleID}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
