import { api, type ApiEnvelope } from "@/shared/api/client";
import type { CreateRolePayload, Permission, Role } from "../model/types";
export const rbacApi = {
  selfAssignableRoles: async () =>
    (await api.get<ApiEnvelope<Role[]>>("/api/v1/roles/self-assignable")).data,
  roles: async () =>
    (
      await api.get<ApiEnvelope<Role[]>>("/api/v1/admin/roles", {
        permission: "roles.read",
      })
    ).data,
  permissions: async () =>
    (
      await api.get<ApiEnvelope<Permission[]>>("/api/v1/admin/permissions", {
        permission: "permissions.read",
      })
    ).data,
  createRole: async (body: CreateRolePayload) =>
    (
      await api.post<ApiEnvelope<Role>>("/api/v1/admin/roles", body, {
        permission: "roles.create",
      })
    ).data,
  assignRole: async (userId: number, roleId: number) =>
    api.post<void>(
      `/api/v1/admin/users/${userId}/roles`,
      { role_id: roleId },
      { permission: "roles.assign" },
    ),
};
