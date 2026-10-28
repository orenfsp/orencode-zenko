import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CompetitionEvent,
  CreateCompetitionEventPayload,
  UpdateCompetitionEventPayload,
} from "../model/types";

function publicPath(competitionId: number) {
  return `/api/v1/competitions/${competitionId}/events`;
}

function adminPath(competitionId: number) {
  return `/api/v1/admin/competitions/${competitionId}/events`;
}

export const competitionEventApi = {
  publicList: async (competitionId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionEvent[]>>(publicPath(competitionId), {
        auth: false,
      })
    ).data,
  publicGet: async (competitionId: number, eventId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionEvent>>(
        `${publicPath(competitionId)}/${eventId}`,
        { auth: false },
      )
    ).data,
  adminList: async (competitionId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionEvent[]>>(adminPath(competitionId), {
        permission: "competition_events.read",
      })
    ).data,
  adminGet: async (competitionId: number, eventId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionEvent>>(
        `${adminPath(competitionId)}/${eventId}`,
        { permission: "competition_events.read" },
      )
    ).data,
  create: async (
    competitionId: number,
    body: CreateCompetitionEventPayload,
  ) =>
    (
      await api.post<ApiEnvelope<CompetitionEvent>>(
        publicPath(competitionId),
        body,
        { permission: "competition_events.create" },
      )
    ).data,
  update: async (
    competitionId: number,
    eventId: number,
    body: UpdateCompetitionEventPayload,
  ) =>
    (
      await api.patch<ApiEnvelope<CompetitionEvent>>(
        `${publicPath(competitionId)}/${eventId}`,
        body,
        { permission: "competition_events.update" },
      )
    ).data,
  remove: (competitionId: number, eventId: number) =>
    api.delete<void>(`${publicPath(competitionId)}/${eventId}`, {
      permission: "competition_events.delete",
    }),
};
