import { api, type ApiEnvelope } from "@/shared/api/client";
import type { Competition, CreateCompetitionPayload, UpdateCompetitionPayload } from "../model/types";

export const competitionApi = {
  create: async (body: CreateCompetitionPayload) =>
    (await api.post<ApiEnvelope<Competition>>("/api/v1/competitions", body, { permission: "competitions.create" })).data,
  update: async (competitionId: number, body: UpdateCompetitionPayload) =>
    (await api.patch<ApiEnvelope<Competition>>(`/api/v1/competitions/${competitionId}`, body, { permission: "competitions.update" })).data,
};
