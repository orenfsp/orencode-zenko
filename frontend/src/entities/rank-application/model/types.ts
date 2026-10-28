export type RankApplicationStatus = "pending" | "approved" | "rejected";
export type RankApplicationReviewStatus = Exclude<RankApplicationStatus, "pending">;

export interface RankApplication {
  id: number;
  athlete_user_id: number;
  rank_id: number;
  discipline_id: number;
  application_type: string;
  competition_id: number | null;
  result_id: number | null;
  submitted_by_user_id: number;
  status: RankApplicationStatus;
  reviewed_by_user_id?: number;
  reviewed_at?: string;
  rejection_reason?: string;
  assigned_at: string | null;
  expires_at: string | null;
  order_number: string | null;
  created_at: string;
}

export interface CreateRankApplicationPayload {
  athlete_user_id?: number;
  rank_id: number;
  discipline_id: number;
  application_type: string;
  competition_id?: number;
  result_id?: number;
}

export interface ReviewRankApplicationPayload {
  status: RankApplicationReviewStatus;
  rejection_reason?: string;
  assigned_at?: string;
  expires_at?: string;
  order_number?: string;
}

export interface RankHistoryItem {
  id: number;
  rank_id: number;
  rank_name: string;
  discipline_id: number;
  discipline_name: string;
  result_id: number | null;
  awarded_at: string | null;
  expires_at: string | null;
  order_number: string | null;
  competition_id: number | null;
  competition_name: string | null;
}

export interface RankApplicationListQuery {
  limit?: number;
  offset?: number;
}
