import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CreateRankApplicationPayload,
  RankApplication,
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
  review: async (applicationId: number, body: ReviewRankApplicationPayload) =>
    (
      await api.patch<ApiEnvelope<RankApplication>>(
        `/api/v1/admin/rank-applications/${applicationId}`,
        body,
        { permission: "roles.assign" },
      )
    ).data,
};
