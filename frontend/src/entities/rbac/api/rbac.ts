import { api, type ApiEnvelope } from "@/shared/api/client";
import type { CreateRolePayload, Permission, Role } from "../model/types";
export const rbacApi = {
  myPermissions: async () =>
    (
      await api.get<ApiEnvelope<string[]>>("/api/v1/my-permission")
    ).data,
  permissionsForActiveRole: async () =>
    (
      await api.get<ApiEnvelope<string[]>>("/api/v1/my-permissions")
    ).data,
  selfAssignableRoles: async () =>
    (await api.get<ApiEnvelope<Role[]>>("/api/v1/roles/self-assignable")).data,
  roles: async () => {
    const roles = (
      await api.get<ApiEnvelope<Array<Partial<Role>>>>("/api/v1/admin/roles", {
        permission: "roles.read",
      })
    ).data;
    return roles.map((role) => ({
      ...role,
      permissions: Array.isArray(role.permissions) ? role.permissions : [],
    })) as Role[];
  },
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
