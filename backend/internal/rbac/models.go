package rbac

const (
	PermissionRolesRead                 = "roles.read"
	PermissionRolesCreate               = "roles.create"
	PermissionRolesAssign               = "roles.assign"
	PermissionPermissionsRead           = "permissions.read"
	PermissionFilesRead                 = "files.read"
	PermissionCompetitionsCreate        = "competitions.create"
	PermissionCompetitionsUpdate        = "competitions.update"
	PermissionCompetitionJudgesManage   = "competition_judges.manage"
	PermissionCompetitionResultsRead    = "competition_results.read"
	PermissionCompetitionResultsManage  = "competition_results.manage"
	PermissionCompetitionResultsApprove = "competition_results.approve"
	PermissionCalendarPlansRead         = "calendar_plans.read"
	PermissionCalendarPlansCreate       = "calendar_plans.create"
	PermissionCalendarPlansUpdate       = "calendar_plans.update"
	PermissionMunicipalitiesRead        = "municipalities.read"
	PermissionMunicipalitiesCreate      = "municipalities.create"

	SuperAdminRoleCode = "super_admin"
)

type PermissionDefinition struct {
	Code        string
	Description string
}

type SeedRoleDefinition struct {
	Code string
	Name string
}

var DefaultSelfAssignableRoles = []SeedRoleDefinition{
	{Code: "participant", Name: "Участник"},
	{Code: "coach", Name: "Тренер"},
	{Code: "judge", Name: "Судья"},
}

var DefaultPermissions = []PermissionDefinition{
	{Code: PermissionRolesRead, Description: "Просмотр ролей"},
	{Code: PermissionRolesCreate, Description: "Создание ролей"},
	{Code: PermissionRolesAssign, Description: "Назначение ролей пользователям"},
	{Code: PermissionPermissionsRead, Description: "Просмотр доступных разрешений"},
	{Code: PermissionFilesRead, Description: "Просмотр документов пользователей"},
	{Code: PermissionCompetitionsCreate, Description: "Создание соревнований"},
	{Code: PermissionCompetitionsUpdate, Description: "Изменение соревнований"},
	{Code: PermissionCompetitionJudgesManage, Description: "Назначение и снятие судей соревнования"},
	{Code: PermissionCompetitionResultsRead, Description: "Просмотр таблиц результатов соревнований"},
	{Code: PermissionCompetitionResultsManage, Description: "Ввод и расчёт результатов соревнований"},
	{Code: PermissionCompetitionResultsApprove, Description: "Утверждение результатов соревнований"},
	{Code: PermissionCalendarPlansRead, Description: "Просмотр календарных планов"},
	{Code: PermissionCalendarPlansCreate, Description: "Создание календарных планов"},
	{Code: PermissionCalendarPlansUpdate, Description: "Изменение календарных планов"},
	{Code: PermissionMunicipalitiesRead, Description: "Просмотр муниципалитетов"},
	{Code: PermissionMunicipalitiesCreate, Description: "Создание муниципалитетов"},
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
