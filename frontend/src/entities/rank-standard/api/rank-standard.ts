import { api, type ApiEnvelope } from "@/shared/api/client";
import type { RankStandard, UpsertRankStandardPayload } from "../model/types";

export const rankStandardApi = {
  list: async () =>
    (await api.get<ApiEnvelope<RankStandard[]>>("/api/v1/admin/rank-standards", {
      permission: "rank_applications.review",
    })).data,
  upsert: async (body: UpsertRankStandardPayload) =>
    (await api.post<ApiEnvelope<RankStandard>>("/api/v1/admin/rank-standards", body, {
      permission: "rank_applications.review",
    })).data,
};
