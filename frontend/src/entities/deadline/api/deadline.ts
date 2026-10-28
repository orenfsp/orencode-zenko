import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CreateReportDeadlinePayload,
  DeadlineItem,
  ReportDeadline,
  UpdateReportDeadlinePayload,
} from "../model/types";

export const deadlineApi = {
  list: async () =>
    (await api.get<ApiEnvelope<DeadlineItem[]>>("/api/v1/deadlines")).data,
  reports: async () =>
    (
      await api.get<ApiEnvelope<ReportDeadline[]>>(
        "/api/v1/admin/report-deadlines",
        { permission: "deadlines.manage" },
      )
    ).data,
  createReport: async (body: CreateReportDeadlinePayload) =>
    (
      await api.post<ApiEnvelope<ReportDeadline>>(
        "/api/v1/admin/report-deadlines",
        body,
        { permission: "deadlines.manage" },
      )
    ).data,
  updateReport: async (deadlineId: number, body: UpdateReportDeadlinePayload) =>
    (
      await api.patch<ApiEnvelope<ReportDeadline>>(
        `/api/v1/admin/report-deadlines/${deadlineId}`,
        body,
        { permission: "deadlines.manage" },
      )
    ).data,
};
