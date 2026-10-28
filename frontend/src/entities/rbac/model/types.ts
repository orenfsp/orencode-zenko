export type PermissionCode =
  | "roles.read"
  | "roles.create"
  | "roles.assign"
  | "permissions.read"
  | "files.read"
  | "calendar_plans.read"
  | "calendar_plans.create"
  | "calendar_plans.update"
  | "competitions.create"
  | "competitions.update"
  | "competition_events.create"
  | "competition_events.update"
  | "competition_events.delete"
  | "competition_events.read"
  | "competition_judges.manage"
  | "competition_results.read"
  | "competition_results.manage"
  | "competition_results.approve"
  | "municipalities.read"
  | "municipalities.create"
  | "anti_doping.review"
  | "dashboard.read"
  | "registry.read"
  | "judge_categories.review"
  | "rank_applications.review"
  | "deadlines.manage";

export interface Permission {
  id: number;
  code: PermissionCode;
  description: string;
}

export interface Role {
  id: number;
  code: string;
  name: string;
  is_system: boolean;
  self_assignable: boolean;
  permissions: string[];
}

export interface CreateRolePayload {
  code: string;
  name: string;
  self_assignable: boolean;
  permission_codes: string[];
}

export interface AssignRolePayload {
  role_id: number;
}
