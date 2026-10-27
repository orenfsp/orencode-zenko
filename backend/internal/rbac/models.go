package rbac

const (
	PermissionRolesRead       = "roles.read"
	PermissionRolesCreate     = "roles.create"
	PermissionRolesAssign     = "roles.assign"
	PermissionPermissionsRead = "permissions.read"
	PermissionFilesRead       = "files.read"

	SuperAdminRoleCode = "super_admin"
)

type PermissionDefinition struct {
	Code        string
	Description string
}

var DefaultPermissions = []PermissionDefinition{
	{Code: PermissionRolesRead, Description: "Просмотр ролей"},
	{Code: PermissionRolesCreate, Description: "Создание ролей"},
	{Code: PermissionRolesAssign, Description: "Назначение ролей пользователям"},
	{Code: PermissionPermissionsRead, Description: "Просмотр доступных разрешений"},
	{Code: PermissionFilesRead, Description: "Просмотр документов пользователей"},
}

type Role struct {
	ID             int64    `json:"id"`
	Code           string   `json:"code"`
	Name           string   `json:"name"`
	IsSystem       bool     `json:"is_system"`
	SelfAssignable bool     `json:"self_assignable"`
	Permissions    []string `json:"permissions"`
}

type Permission struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type CreateRoleRequest struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	SelfAssignable  bool     `json:"self_assignable"`
	PermissionCodes []string `json:"permission_codes"`
}

type AssignRoleRequest struct {
	RoleID int64 `json:"role_id"`
}

type Identity struct {
	UserID int64
	RoleID *int64
}
