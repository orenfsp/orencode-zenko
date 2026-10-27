package rbac

import (
	"context"

	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

// RoleChecker exposes the role lookup needed by modules whose rules depend on
// role codes rather than permissions.
type RoleChecker interface {
	HasRole(ctx context.Context, userID int64, roleCode string) (bool, error)
}

type sqlRoleChecker struct {
	queries *db.Queries
}

func NewRoleChecker(queries *db.Queries) RoleChecker {
	return &sqlRoleChecker{queries: queries}
}

func (c *sqlRoleChecker) HasRole(ctx context.Context, userID int64, roleCode string) (bool, error) {
	return c.queries.HasRole(ctx, db.HasRoleParams{UserID: userID, Code: roleCode})
}
