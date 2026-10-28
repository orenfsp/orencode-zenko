export interface CompetitionJudgeUser {
  id: number;
  first_name: string;
  last_name: string;
  middle_name: string | null;
}

export interface CompetitionJudgePosition {
  id: number;
  name: string;
}

export interface CompetitionJudge {
  id: number;
  competition_id: number;
  judge: CompetitionJudgeUser;
  position: CompetitionJudgePosition;
  appointed_at: string;
}

export interface AssignCompetitionJudgePayload {
  judge_user_id: number;
  position_id: number;
}

export interface CreatedCompetitionJudge {
  id: number;
  competition_id: number;
  judge_user_id: number;
  position_id: number;
  appointed_at: string;
}

export interface JudgeCompetitionReference {
  id: number;
  name: string;
  discipline: { id: number; name: string };
  competition_type: { id: number; name: string };
  format: string;
  start_at: string;
  end_at: string;
  status: string;
  position?: CompetitionJudgePosition;
  appointed_at?: string;
}

export interface JudgeAssignment {
  id: number;
  competition: JudgeCompetitionReference;
  position: CompetitionJudgePosition;
  appointed_at: string;
}

export interface JudgeProfileQuery {
  limit?: number;
  offset?: number;
}
