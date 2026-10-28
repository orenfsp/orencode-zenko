import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CalendarCompetition,
  Competition,
  CompetitionCalendarQuery,
  CreateCompetitionPayload,
  PaginationQuery,
  UpdateCompetitionPayload,
} from "../model/types";

export const competitionApi = {
  list: async (query: PaginationQuery = {}) =>
    (
      await api.get<ApiEnvelope<Competition[]>>("/api/v1/competitions", {
        auth: false,
        params: { ...query },
      })
    ).data,
  get: async (competitionId: number) =>
    (
      await api.get<ApiEnvelope<Competition>>(
        `/api/v1/competitions/${competitionId}`,
        { auth: false },
      )
    ).data,
  calendar: async (query: CompetitionCalendarQuery) =>
    (
      await api.get<ApiEnvelope<CalendarCompetition[]>>(
        "/api/v1/calendar/competitions",
        { auth: false, params: { ...query } },
      )
    ).data,
  create: async (body: CreateCompetitionPayload) =>
    (await api.post<ApiEnvelope<Competition>>("/api/v1/competitions", body, { permission: "competitions.create" })).data,
  update: async (competitionId: number, body: UpdateCompetitionPayload) =>
    (await api.patch<ApiEnvelope<Competition>>(`/api/v1/competitions/${competitionId}`, body, { permission: "competitions.update" })).data,
};
