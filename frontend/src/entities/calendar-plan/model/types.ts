export interface CalendarPlan {
  id: number;
  name: string;
  year: number;
}

export interface CalendarPlanCompetition {
  id: number;
  name: string;
  start_at: string;
  end_at: string;
  status: string;
  discipline: { id: number; name: string };
  calendar_plan_number: string | null;
}

export interface CalendarPlanListQuery {
  year?: number;
  limit?: number;
  offset?: number;
}

export interface PaginationQuery {
  limit?: number;
  offset?: number;
}

export interface CreateCalendarPlanPayload {
  name: string;
  year: number;
}

export type UpdateCalendarPlanPayload = Partial<CreateCalendarPlanPayload>;
