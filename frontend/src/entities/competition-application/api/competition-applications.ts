import { api, type ApiEnvelope } from "@/shared/api/client";
import type { CompetitionApplication, CompetitionApplicationReviewQuery } from "../model/types";

export const competitionApplicationApi = {
  create: async (body: { event_id: number; athlete_user_id?: number }) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>("/api/v1/competition-applications", body)).data,
  mine: async () =>
    (await api.get<ApiEnvelope<CompetitionApplication[]>>("/api/v1/competition-applications/my")).data,
  get: async (id: number) =>
    (await api.get<ApiEnvelope<CompetitionApplication>>(`/api/v1/competition-applications/${id}`)).data,
  withdraw: async (id: number) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>(`/api/v1/competition-applications/${id}/withdraw`)).data,
  pending: async (query: CompetitionApplicationReviewQuery = {}) =>
    (await api.get<ApiEnvelope<CompetitionApplication[]>>("/api/v1/admin/competition-applications", { permission: "competition_applications.review", params: { ...query } })).data,
  getForReview: async (id: number) =>
    (await api.get<ApiEnvelope<CompetitionApplication>>(`/api/v1/admin/competition-applications/${id}`, { permission: "competition_applications.review" })).data,
  approve: async (id: number) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>(`/api/v1/admin/competition-applications/${id}/approve`, undefined, { permission: "competition_applications.review" })).data,
  reject: async (id: number) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>(`/api/v1/admin/competition-applications/${id}/reject`, undefined, { permission: "competition_applications.review" })).data,
};
