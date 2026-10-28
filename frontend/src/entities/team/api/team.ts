import { api, type ApiEnvelope } from "@/shared/api/client";
import type { CreateTeamPayload, Team } from "../model/types";

export const teamApi = {
  list: async () =>
    (await api.get<ApiEnvelope<Team[]>>("/api/v1/admin/teams", { permission: "teams.manage" })).data,
  create: async (body: CreateTeamPayload) =>
    (await api.post<ApiEnvelope<Team>>("/api/v1/admin/teams", body, { permission: "teams.manage" })).data,
};
