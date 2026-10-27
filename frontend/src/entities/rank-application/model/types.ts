export type RankApplicationStatus = "pending" | "approved" | "rejected";
export type RankApplicationReviewStatus = Exclude<RankApplicationStatus, "pending">;

export interface RankApplication {
  id: number;
  athlete_user_id: number;
  rank_id: number;
  discipline_id: number;
  submitted_by_user_id: number;
  status: RankApplicationStatus;
  reviewed_by_user_id?: number;
  reviewed_at?: string;
  rejection_reason?: string;
  created_at: string;
}

export interface CreateRankApplicationPayload {
  athlete_user_id?: number;
  rank_id: number;
  discipline_id: number;
}

export type ReviewRankApplicationPayload =
  | { status: "approved"; rejection_reason?: never }
  | { status: "rejected"; rejection_reason: string };
