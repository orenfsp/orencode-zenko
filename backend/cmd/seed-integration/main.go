package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
	"github.com/orenfsp/orencode-zenko.git/internal/integration"
	"github.com/orenfsp/orencode-zenko.git/pkg/password"
)

var clientKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,100}$`)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	clientKey := strings.TrimSpace(os.Getenv("INTEGRATION_CLIENT_ID"))
	clientSecret := os.Getenv("INTEGRATION_CLIENT_SECRET")
	name := strings.TrimSpace(os.Getenv("INTEGRATION_CLIENT_NAME"))
	ownerEmail := strings.TrimSpace(os.Getenv("INTEGRATION_OWNER_EMAIL"))
	permissions, err := parsePermissions(os.Getenv("INTEGRATION_PERMISSIONS"))
	if err != nil {
		log.Fatal(err)
	}
	if databaseURL == "" || !clientKeyPattern.MatchString(clientKey) || len(clientSecret) < 16 || len(clientSecret) > 72 ||
		name == "" || len(name) > 200 || ownerEmail == "" {
		log.Fatal("DATABASE_URL, valid INTEGRATION_CLIENT_ID, 16-72 byte INTEGRATION_CLIENT_SECRET, INTEGRATION_CLIENT_NAME and INTEGRATION_OWNER_EMAIL are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := seed(ctx, pool, clientKey, clientSecret, name, ownerEmail, permissions); err != nil {
		log.Fatal(err)
	}
	log.Printf("integration client %q seeded with permissions: %s", clientKey, strings.Join(permissions, ","))
}

func seed(
	ctx context.Context, pool *pgxpool.Pool, clientKey, clientSecret, name, ownerEmail string, permissions []string,
) error {
	secretHash, err := password.Hash(clientSecret)
	if err != nil {
		return err
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := db.New(pool).WithTx(tx)
	ownerID, err := queries.GetIntegrationOwnerIDByEmail(ctx, ownerEmail)
	if err != nil {
		return fmt.Errorf("active integration owner not found: %w", err)
	}
	clientID, err := queries.UpsertIntegrationClient(ctx, db.UpsertIntegrationClientParams{
		ClientKey: clientKey, Name: name, SecretHash: secretHash, OwnerUserID: ownerID,
	})
	if err != nil {
		return err
	}
	if err := queries.ClearIntegrationClientPermissions(ctx, clientID); err != nil {
		return err
	}
	for _, permission := range permissions {
		if err := queries.GrantIntegrationClientPermission(ctx, db.GrantIntegrationClientPermissionParams{
			ClientID: clientID, PermissionCode: permission,
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func parsePermissions(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		raw = integration.PermissionResultsImport + "," + integration.PermissionWebhooksManage
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, 2)
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimSpace(value)
		if _, ok := integration.AllowedPermissions[value]; !ok {
			return nil, fmt.Errorf("unsupported integration permission %q", value)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}
