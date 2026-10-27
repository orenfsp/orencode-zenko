import { api, type ApiEnvelope } from "@/shared/api/client";
import type {
  CalendarPlan,
  CalendarPlanCompetition,
  CalendarPlanListQuery,
  CreateCalendarPlanPayload,
  PaginationQuery,
  UpdateCalendarPlanPayload,
} from "../model/types";

export const calendarPlanApi = {
  list: async (query: CalendarPlanListQuery = {}) =>
    (
      await api.get<ApiEnvelope<CalendarPlan[]>>(
        "/api/v1/admin/calendar-plans",
        {
          params: {
            year: query.year,
            limit: query.limit,
            offset: query.offset,
          },
          permission: "calendar_plans.read",
        },
      )
    ).data,
  get: async (planId: number) =>
    (
      await api.get<ApiEnvelope<CalendarPlan>>(
        `/api/v1/admin/calendar-plans/${planId}`,
        { permission: "calendar_plans.read" },
      )
    ).data,
  create: async (body: CreateCalendarPlanPayload) =>
    (
      await api.post<ApiEnvelope<CalendarPlan>>(
        "/api/v1/admin/calendar-plans",
        body,
        { permission: "calendar_plans.create" },
      )
    ).data,
  update: async (planId: number, body: UpdateCalendarPlanPayload) =>
    (
      await api.patch<ApiEnvelope<CalendarPlan>>(
        `/api/v1/admin/calendar-plans/${planId}`,
        body,
        { permission: "calendar_plans.update" },
      )
    ).data,
  competitions: async (planId: number, query: PaginationQuery = {}) =>
    (
      await api.get<ApiEnvelope<CalendarPlanCompetition[]>>(
        `/api/v1/admin/calendar-plans/${planId}/competitions`,
        {
          params: { limit: query.limit, offset: query.offset },
          permission: "calendar_plans.read",
        },
      )
    ).data,
};
