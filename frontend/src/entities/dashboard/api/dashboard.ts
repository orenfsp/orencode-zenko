import { api, type ApiEnvelope } from "@/shared/api/client";
import type { DashboardData } from "../model/types";

export const dashboardApi = {
  get: async (year?: number) =>
    (
      await api.get<ApiEnvelope<DashboardData>>("/api/v1/admin/dashboard", {
        permission: "dashboard.read",
        params: year ? { year } : undefined,
      })
    ).data,
};
