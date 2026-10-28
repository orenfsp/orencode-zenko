export type CompetitionApplicationStatus = "pending" | "approved" | "rejected" | "withdrawn";

export interface CompetitionApplication {
  id: number;
  event_id: number;
  athlete_user_id: number;
  submitted_by_user_id: number;
  status: CompetitionApplicationStatus;
  submitted_at: string;
  competition_id: number;
  competition_name: string;
  competition_start_at: string;
  competition_end_at: string;
  registration_starts_at: string;
  registration_ends_at: string;
  discipline_id: number;
  discipline_name: string;
  age_group_id?: number | null;
  age_group_name?: string | null;
  athlete_first_name: string;
  athlete_last_name: string;
  athlete_email?: string;
  submitter_first_name?: string;
  submitter_last_name?: string;
  can_withdraw?: boolean;
  reviewed_by_user_id?: number;
  reviewed_at?: string;
}

export interface CompetitionApplicationReviewQuery {
  event_id?: number;
  status?: CompetitionApplicationStatus;
  limit?: number;
  offset?: number;
}
