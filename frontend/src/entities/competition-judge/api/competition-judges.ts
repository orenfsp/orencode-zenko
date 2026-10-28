import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  AssignCompetitionJudgePayload,
  CompetitionJudge,
  CreatedCompetitionJudge,
  JudgeAssignment,
  JudgeCompetitionReference,
  JudgeProfileQuery,
} from "../model/types";

export const competitionJudgeApi = {
  profileCompetitions: async (query: JudgeProfileQuery = {}) =>
    (
      await api.get<ApiEnvelope<JudgeCompetitionReference[]>>(
        "/api/v1/profile/competitions",
        { params: { ...query } },
      )
    ).data,
  profileAssignments: async (query: JudgeProfileQuery = {}) =>
    (
      await api.get<ApiEnvelope<JudgeAssignment[]>>(
        "/api/v1/profile/assignments",
        { params: { ...query } },
      )
    ).data,
  publicList: async (competitionId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionJudge[]>>(
        `/api/v1/competitions/${competitionId}/judges`,
        { auth: false },
      )
    ).data,
  adminList: async (competitionId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionJudge[]>>(
        `/api/v1/admin/competitions/${competitionId}/judges`,
        { permission: "competition_judges.manage" },
      )
    ).data,
  assign: async (
    competitionId: number,
    body: AssignCompetitionJudgePayload,
  ) =>
    (
      await api.post<ApiEnvelope<CreatedCompetitionJudge>>(
        `/api/v1/competitions/${competitionId}/judges`,
        body,
        { permission: "competition_judges.manage" },
      )
    ).data,
  remove: (competitionId: number, assignmentId: number) =>
    api.delete<void>(
      `/api/v1/competitions/${competitionId}/judges/${assignmentId}`,
      { permission: "competition_judges.manage" },
    ),
};
