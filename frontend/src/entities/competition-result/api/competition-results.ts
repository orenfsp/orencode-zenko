import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CompetitionProtocol,
  CompetitionResultApproval,
  CompetitionResultHistoryItem,
  CompetitionResultTable,
  CreateResultTemplatePayload,
  ResultSettingsPayload,
  SaveManualResultsPayload,
  SaveResultsPayload,
  ResultTemplate,
  UpdateResultTemplatePayload,
} from "../model/types";

function resultsPath(competitionId: number, eventId: number) {
  return `/api/v1/admin/competitions/${competitionId}/events/${eventId}/results`;
}

function protocolFilename(disposition: string | null, eventId: number) {
  return disposition?.match(/filename="([^"]+)"/i)?.[1] ?? `protocol-${eventId}.pdf`;
}

export const competitionResultApi = {
  templates: async (disciplineId?: number) =>
    (
      await api.get<ApiEnvelope<ResultTemplate[]>>("/api/v1/admin/result-templates", {
        permission: "competition_results.read",
        params: disciplineId ? { discipline_id: disciplineId } : undefined,
      })
    ).data,
  createTemplate: async (body: CreateResultTemplatePayload) =>
    (
      await api.post<ApiEnvelope<ResultTemplate>>("/api/v1/admin/result-templates", body, {
        permission: "competition_results.manage",
      })
    ).data,
  updateTemplate: async (templateId: number, body: UpdateResultTemplatePayload) =>
    (
      await api.patch<ApiEnvelope<ResultTemplate>>(`/api/v1/admin/result-templates/${templateId}`, body, {
        permission: "competition_results.manage",
      })
    ).data,
  get: async (competitionId: number, eventId: number) =>
    (
      await api.get<ApiEnvelope<CompetitionResultTable>>(
        resultsPath(competitionId, eventId),
        { permission: "competition_results.read" },
      )
    ).data,
  configure: async (
    competitionId: number,
    eventId: number,
    body: ResultSettingsPayload,
  ) =>
    (
      await api.put<ApiEnvelope<CompetitionResultTable>>(
        `${resultsPath(competitionId, eventId)}/settings`,
        body,
        { permission: "competition_results.manage" },
      )
    ).data,
  save: async (
    competitionId: number,
    eventId: number,
    body: SaveResultsPayload,
  ) =>
    (
      await api.put<ApiEnvelope<CompetitionResultTable>>(
        resultsPath(competitionId, eventId),
        body,
        { permission: "competition_results.manage" },
      )
    ).data,
  saveManual: async (
    competitionId: number,
    eventId: number,
    body: SaveManualResultsPayload,
  ) =>
    (
      await api.put<ApiEnvelope<CompetitionResultTable>>(
        `${resultsPath(competitionId, eventId)}/manual`,
        body,
        { permission: "competition_results.manage" },
      )
    ).data,
  calculate: async (competitionId: number, eventId: number) =>
    (
      await api.post<ApiEnvelope<CompetitionResultTable>>(
        `${resultsPath(competitionId, eventId)}/calculate`,
        undefined,
        { permission: "competition_results.manage" },
      )
    ).data,
  approve: async (competitionId: number, eventId: number) =>
    (
      await api.post<ApiEnvelope<CompetitionResultApproval>>(
        `${resultsPath(competitionId, eventId)}/approve`,
        undefined,
        { permission: "competition_results.approve" },
      )
    ).data,
  history: async (query: { athlete_user_id?: number; discipline_id?: number } = {}) =>
    (
      await api.get<ApiEnvelope<CompetitionResultHistoryItem[]>>(
        "/api/v1/profile/competition-results",
        { params: { ...query } },
      )
    ).data,
  async protocol(
    competitionId: number,
    eventId: number,
  ): Promise<CompetitionProtocol> {
    const response = await api.get<Response>(
      `/api/v1/competitions/${competitionId}/events/${eventId}/protocol`,
      { auth: false, responseType: "response" },
    );
    return {
      blob: await response.blob(),
      filename: protocolFilename(
        response.headers.get("Content-Disposition"),
        eventId,
      ),
    };
  },
  uploadProtocol: async (competitionId: number, eventId: number, file: File) => {
    const body = new FormData();
    body.append("file", file);
    return (
      await api.put<ApiEnvelope<{ protocol_url: string }>>(
        `/api/v1/admin/competitions/${competitionId}/events/${eventId}/protocol`,
        body,
        { permission: "competition_results.approve" },
      )
    ).data;
  },
  removeProtocol: (competitionId: number, eventId: number) =>
    api.delete<void>(
      `/api/v1/admin/competitions/${competitionId}/events/${eventId}/protocol`,
      { permission: "competition_results.approve" },
    ),
};
