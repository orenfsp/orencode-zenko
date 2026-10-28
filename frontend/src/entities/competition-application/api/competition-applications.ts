import { api, type ApiEnvelope } from "@/shared/api/client";
import type { CompetitionApplication, CompetitionApplicationListQuery, CreateCompetitionApplicationPayload } from "../model/types";

export const competitionApplicationApi = {
  create: async (body: CreateCompetitionApplicationPayload) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>("/api/v1/competition-applications", body)).data,
  mine: async () =>
    (await api.get<ApiEnvelope<CompetitionApplication[]>>("/api/v1/competition-applications/my")).data,
  get: async (applicationId: number) =>
    (await api.get<ApiEnvelope<CompetitionApplication>>(`/api/v1/competition-applications/${applicationId}`)).data,
  withdraw: async (applicationId: number) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>(`/api/v1/competition-applications/${applicationId}/withdraw`)).data,
  reviewList: async (query: CompetitionApplicationListQuery = {}) =>
    (await api.get<ApiEnvelope<CompetitionApplication[]>>("/api/v1/admin/competition-applications", {
      permission: "competition_applications.review",
      params: { ...query },
    })).data,
  reviewGet: async (applicationId: number) =>
    (await api.get<ApiEnvelope<CompetitionApplication>>(`/api/v1/admin/competition-applications/${applicationId}`, {
      permission: "competition_applications.review",
    })).data,
  approve: async (applicationId: number) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>(`/api/v1/admin/competition-applications/${applicationId}/approve`, undefined, {
      permission: "competition_applications.review",
    })).data,
  reject: async (applicationId: number) =>
    (await api.post<ApiEnvelope<CompetitionApplication>>(`/api/v1/admin/competition-applications/${applicationId}/reject`, undefined, {
      permission: "competition_applications.review",
    })).data,
};

