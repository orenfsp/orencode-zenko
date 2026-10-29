import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CreateJudgeCategoryApplicationPayload,
  CurrentJudgeCategory,
  JudgeCategory,
  JudgeCategoryApplication,
  JudgeCategoryApplicationQuery,
  JudgeCompetitionAssignment,
  ReviewJudgeCategoryApplicationPayload,
} from "../model/types";

export const judgeCategoryApi = {
  categories: async () =>
    (
      await api.get<ApiEnvelope<JudgeCategory[]>>("/api/v1/judge-categories", {
        auth: false,
      })
    ).data,
  create: async (body: CreateJudgeCategoryApplicationPayload) =>
    (
      await api.post<ApiEnvelope<JudgeCategoryApplication>>(
        "/api/v1/judge-category-applications",
        body,
      )
    ).data,
  mine: async () =>
    (
      await api.get<ApiEnvelope<JudgeCategoryApplication[]>>(
        "/api/v1/judge-category-applications/mine",
      )
    ).data,
  current: async () =>
    (
      await api.get<ApiEnvelope<CurrentJudgeCategory | null>>(
        "/api/v1/judge-category/current",
      )
    ).data,
  assignments: async () =>
    (
      await api.get<ApiEnvelope<JudgeCompetitionAssignment[]>>(
        "/api/v1/judge/competitions",
      )
    ).data,
  pending: async (query: JudgeCategoryApplicationQuery = {}) =>
    (
      await api.get<ApiEnvelope<JudgeCategoryApplication[]>>(
        "/api/v1/admin/judge-category-applications",
        { permission: "judge_categories.review", params: { ...query } },
      )
    ).data,
  review: async (
    applicationId: number,
    body: ReviewJudgeCategoryApplicationPayload,
  ) =>
    (
      await api.patch<ApiEnvelope<JudgeCategoryApplication>>(
        `/api/v1/admin/judge-category-applications/${applicationId}`,
        body,
        { permission: "judge_categories.review" },
      )
    ).data,
};
