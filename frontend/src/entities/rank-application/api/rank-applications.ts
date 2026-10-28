import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CreateRankApplicationPayload,
  RankApplication,
  RankApplicationListQuery,
  RankHistoryItem,
  ReviewRankApplicationPayload,
} from "../model/types";

export const rankApplicationApi = {
  create: async (body: CreateRankApplicationPayload) =>
    (
      await api.post<ApiEnvelope<RankApplication>>(
        "/api/v1/rank-applications",
        body,
      )
    ).data,
  mine: async () =>
    (
      await api.get<ApiEnvelope<RankApplication[]>>(
        "/api/v1/rank-applications/mine",
      )
    ).data,
  history: async () =>
    (
      await api.get<ApiEnvelope<RankHistoryItem[]>>("/api/v1/ranks/history")
    ).data,
  pending: async (query: RankApplicationListQuery = {}) =>
    (
      await api.get<ApiEnvelope<RankApplication[]>>(
        "/api/v1/admin/rank-applications",
        { permission: "rank_applications.review", params: { ...query } },
      )
    ).data,
  review: async (applicationId: number, body: ReviewRankApplicationPayload) =>
    (
      await api.patch<ApiEnvelope<RankApplication>>(
        `/api/v1/admin/rank-applications/${applicationId}`,
        body,
        { permission: "rank_applications.review" },
      )
    ).data,
};
