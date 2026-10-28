package rbac

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *db.Queries, pool *pgxpool.Pool) Repository {
	return &sqlRepository{queries: queries, pool: pool}
}

func (r *sqlRepository) HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	return r.queries.HasPermission(ctx, db.HasPermissionParams{ID: userID, Code: permissionCode})
}

func (r *sqlRepository) CreateRole(ctx context.Context, request CreateRoleRequest) (Role, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Role{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	row, err := queries.CreateRole(ctx, db.CreateRoleParams{
		Code:           request.Code,
		Name:           request.Name,
		SelfAssignable: request.SelfAssignable,
	})
	if err != nil {
		return Role{}, mapPostgresError(err)
	}
	role := Role{
		ID:             row.ID,
		Code:           row.Code,
		Name:           row.Name,
		IsSystem:       row.IsSystem,
		SelfAssignable: row.SelfAssignable,
		Permissions:    request.PermissionCodes,
	}

	for _, permissionCode := range request.PermissionCodes {
		permissionID, err := queries.GetPermissionIDByCode(ctx, permissionCode)
		if errors.Is(err, pgx.ErrNoRows) {
			return Role{}, ErrPermissionNotFound
		}
		if err != nil {
			return Role{}, err
		}
		if err := queries.AssignRolePermission(ctx, db.AssignRolePermissionParams{
			RoleID:       role.ID,
			PermissionID: permissionID,
		}); err != nil {
			return Role{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Role{}, err
	}
	return role, nil
}

func (r *sqlRepository) AssignRole(ctx context.Context, userID, roleID int64) error {
	userExists, err := r.queries.UserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !userExists {
		return ErrUserNotFound
	}
	roleExists, err := r.queries.RoleExists(ctx, roleID)
	if err != nil {
		return err
	}
	if !roleExists {
		return ErrRoleNotFound
	}
	return r.queries.AssignRole(ctx, db.AssignRoleParams{UserID: userID, RoleID: roleID})
}

func (r *sqlRepository) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.queries.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]Role, 0, len(rows))
	for _, row := range rows {
		permissionCodes, err := r.queries.ListPermissionCodesByRole(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, Role{
			ID:             row.ID,
			Code:           row.Code,
			Name:           row.Name,
			IsSystem:       row.IsSystem,
			SelfAssignable: row.SelfAssignable,
			Permissions:    permissionCodes,
		})
	}
	return roles, nil
}

func (r *sqlRepository) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := r.queries.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	permissions := make([]Permission, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, Permission{
			ID:          row.ID,
			Code:        row.Code,
			Description: row.Description,
		})
	}
	return permissions, nil
}

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	if pgErr.Code == "23505" && (pgErr.ConstraintName == "role_code_key" || pgErr.ConstraintName == "role_name_key") {
		return ErrRoleAlreadyExists
	}
	return err
}
func (r *sqlRepository) ListSelfAssignableRoles(ctx context.Context) ([]Role, error) {
	rows, err := r.queries.ListSelfAssignableRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]Role, 0, len(rows))
	for _, row := range rows {
		codes, err := r.queries.ListPermissionCodesByRole(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, Role{ID: row.ID, Code: row.Code, Name: row.Name,
			IsSystem: row.IsSystem, SelfAssignable: row.SelfAssignable, Permissions: codes})
	}
	return roles, nil
}

func (r *sqlRepository) ListRolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	return r.queries.ListPermissionCodesByRoleID(ctx, roleID)
}
