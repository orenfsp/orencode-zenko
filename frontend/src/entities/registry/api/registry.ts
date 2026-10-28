import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  RegistryOrganization,
  RegistryOrganizationQuery,
  RegistryPage,
  RegistryUser,
  RegistryUserQuery,
} from "../model/types";

export const registryApi = {
  report: async () => {
    const response = await api.get<Response>("/api/v1/admin/reports/users.xlsx", {
      permission: "registry.read",
      responseType: "response",
    });
    return {
      blob: await response.blob(),
      filename: response.headers.get("Content-Disposition")?.match(/filename="([^"]+)"/)?.[1] ?? "federation-users.xlsx",
      fileId: response.headers.get("X-File-ID"),
    };
  },
  users: async (query: RegistryUserQuery = {}) =>
    (
      await api.get<ApiEnvelope<RegistryPage<RegistryUser>>>(
        "/api/v1/admin/registry/users",
        { permission: "registry.read", params: { ...query } },
      )
    ).data,
  organizations: async (query: RegistryOrganizationQuery = {}) =>
    (
      await api.get<ApiEnvelope<RegistryPage<RegistryOrganization>>>(
        "/api/v1/admin/registry/organizations",
        { permission: "registry.read", params: { ...query } },
      )
    ).data,
};
