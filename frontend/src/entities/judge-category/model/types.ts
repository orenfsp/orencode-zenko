export type JudgeCategoryApplicationStatus = "pending" | "approved" | "rejected";

export interface JudgeCategoryApplication {
  id: number;
  judge_user_id: number;
  judge_name?: string;
  category_id: number;
  category_name?: string;
  application_type: string;
  status: JudgeCategoryApplicationStatus;
  submitted_at: string;
  reviewed_by_user_id: number | null;
  reviewed_at: string | null;
  assigned_at: string | null;
  expires_at: string | null;
  order_number: string | null;
  rejection_reason: string | null;
}

export interface CurrentJudgeCategory {
  id: number;
  user_id: number;
  category_id: number;
  category_name: string;
  assigned_at: string;
  expires_at: string | null;
  order_number: string | null;
}

export interface JudgeCategory {
  id: number;
  name: string;
}

export interface JudgeCompetitionAssignment {
  id: number;
  competition_id: number;
  competition_name: string;
  start_at: string;
  end_at: string;
  status: string;
  position_id: number;
  position_name: string;
  appointed_at: string;
}

export interface CreateJudgeCategoryApplicationPayload {
  category_id: number;
  application_type: string;
}

export interface ReviewJudgeCategoryApplicationPayload {
  status: Exclude<JudgeCategoryApplicationStatus, "pending">;
  assigned_at?: string;
  expires_at?: string;
  order_number?: string;
  rejection_reason?: string;
}

export interface JudgeCategoryApplicationQuery {
  limit?: number;
  offset?: number;
}
