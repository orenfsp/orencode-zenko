import { api, type ApiEnvelope } from "@/shared/api/client";
import type { CreateRankPayload, Rank } from "../model/types";

export const rankApi = {
  create: async (body: CreateRankPayload) =>
    (
      await api.post<ApiEnvelope<Rank>>("/api/v1/admin/rank", body, {
        permission: "rank_applications.review",
      })
    ).data,
  list: async () =>
    (await api.get<ApiEnvelope<Rank[]>>("/api/v1/rank", { auth: false })).data,
  get: async (rankId: number) =>
    (
      await api.get<ApiEnvelope<Rank>>(`/api/v1/rank/${rankId}`, {
        auth: false,
      })
    ).data,
};
