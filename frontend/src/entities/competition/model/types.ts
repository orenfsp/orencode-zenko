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
