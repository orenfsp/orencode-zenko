export interface CompetitionReference {
  id: number;
  name: string;
}

export interface CompetitionLevelReference {
  id: number;
  code: string | null;
  name: string | null;
}

export interface CompetitionCalendarPlanReference {
  id: number;
  name: string | null;
  year: number | null;
  number: string | null;
}

export interface Competition {
  id: number;
  name: string;
  description: string | null;
  discipline: CompetitionReference;
  competition_type: CompetitionReference;
  competition_level: CompetitionLevelReference;
  format: string;
  start_at: string;
  end_at: string;
  registration_starts_at: string;
  registration_ends_at: string;
  venue_address: string | null;
  online_url: string | null;
  status: string;
  calendar_plan: CompetitionCalendarPlanReference | null;
  regulation_url: string | null;
}

export interface CreateCompetitionPayload {
  name: string;
  description: string | null;
  discipline_id: number;
  competition_type_id: number;
  competition_level_id: number;
  format: string;
  start_at: string;
  end_at: string;
  registration_starts_at: string;
  registration_ends_at: string;
  venue_address: string | null;
  online_url?: string | null;
  calendar_plan_id?: number | null;
  calendar_plan_number?: string | null;
  regulation_url?: string | null;
}

export type UpdateCompetitionPayload = Partial<CreateCompetitionPayload>;

export interface PaginationQuery {
  limit?: number;
  offset?: number;
}

export interface CompetitionCalendarQuery extends PaginationQuery {
  from: string;
  to: string;
  discipline_id?: number;
  level_id?: number;
}

export type CompetitionAdminStatus =
  | "draft"
  | "under_review"
  | "published"
  | "cancelled";

export interface CompetitionAdminQuery extends PaginationQuery {
  status?: CompetitionAdminStatus;
}

export interface CompetitionHistoryEntry {
  id: number;
  competition_id: number;
  from_status: string | null;
  to_status: string;
  changed_by_user_id: number | null;
  changed_at: string;
}

export interface CalendarCompetition {
  id: number;
  name: string;
  discipline: CompetitionReference;
  competition_type: CompetitionReference;
  competition_level: CompetitionLevelReference;
  format: string;
  start_at: string;
  end_at: string;
  venue_address: string | null;
  online_url: string | null;
  status: string;
}
