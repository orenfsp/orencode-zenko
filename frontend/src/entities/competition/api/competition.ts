import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CalendarCompetition,
  Competition,
  CompetitionCalendarQuery,
  CompetitionAdminQuery,
  CompetitionHistoryEntry,
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
  adminList: async (query: CompetitionAdminQuery = {}) =>
    (
      await api.get<ApiEnvelope<Competition[]>>("/api/v1/admin/competitions", {
        permission: "competitions.update",
        params: { ...query },
      })
    ).data,
  adminGet: async (competitionId: number) =>
    (
      await api.get<ApiEnvelope<Competition>>(
        `/api/v1/admin/competitions/${competitionId}`,
        { permission: "competitions.update" },
      )
    ).data,
  submit: async (competitionId: number) =>
    (
      await api.post<ApiEnvelope<Competition>>(
        `/api/v1/admin/competitions/${competitionId}/submit`,
        undefined,
        { permission: "competitions.update" },
      )
    ).data,
  publish: async (competitionId: number) =>
    (
      await api.post<ApiEnvelope<Competition>>(
        `/api/v1/admin/competitions/${competitionId}/publish`,
        undefined,
        { permission: "competitions.update" },
      )
    ).data,
  openRegistration: async (competitionId: number) =>
    (await api.post<ApiEnvelope<Competition>>(`/api/v1/admin/competitions/${competitionId}/open-registration`, undefined, { permission: "competitions.update" })).data,
  closeRegistration: async (competitionId: number) =>
    (await api.post<ApiEnvelope<Competition>>(`/api/v1/admin/competitions/${competitionId}/close-registration`, undefined, { permission: "competitions.update" })).data,
  start: async (competitionId: number) =>
    (await api.post<ApiEnvelope<Competition>>(`/api/v1/admin/competitions/${competitionId}/start`, undefined, { permission: "competitions.update" })).data,
  startResults: async (competitionId: number) =>
    (await api.post<ApiEnvelope<Competition>>(`/api/v1/admin/competitions/${competitionId}/start-results`, undefined, { permission: "competitions.update" })).data,
  complete: async (competitionId: number) =>
    (await api.post<ApiEnvelope<Competition>>(`/api/v1/admin/competitions/${competitionId}/complete`, undefined, { permission: "competitions.update" })).data,
  cancel: async (competitionId: number) =>
    (
      await api.post<ApiEnvelope<Competition>>(
        `/api/v1/admin/competitions/${competitionId}/cancel`,
        undefined,
        { permission: "competitions.update" },
      )
    ).data,
  returnToDraft: async (competitionId: number) =>
    (
      await api.post<ApiEnvelope<Competition>>(
        `/api/v1/admin/competitions/${competitionId}/return`,
        undefined,
        { permission: "competitions.update" },
      )
    ).data,
  history: async (competitionId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionHistoryEntry[]>>(
        `/api/v1/admin/competitions/${competitionId}/history`,
        { permission: "competitions.update" },
      )
    ).data,
};
