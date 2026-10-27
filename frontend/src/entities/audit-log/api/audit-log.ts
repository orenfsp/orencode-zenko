import { api, type ApiEnvelope } from "@/shared/api/client";
import type { AuditLogEntry, AuditLogQuery } from "../model/types";

export const auditLogApi = {
  list: async (params: AuditLogQuery = {}) =>
    (await api.get<ApiEnvelope<AuditLogEntry[]>>("/api/v1/admin/audit-logs", {
      params: { limit: params.limit, offset: params.offset },
      permission: "roles.read",
    })).data,
};
