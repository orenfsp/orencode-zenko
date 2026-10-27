package rbac

import "context"

type Repository interface {
	HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
	CreateRole(ctx context.Context, request CreateRoleRequest) (Role, error)
	AssignRole(ctx context.Context, userID, roleID int64) error
	ListRoles(ctx context.Context) ([]Role, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
}
