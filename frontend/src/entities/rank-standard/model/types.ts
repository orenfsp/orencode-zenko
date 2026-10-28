export interface RankStandard {
  id: number;
  rank_id: number;
  discipline_id: number;
  minimum_place?: number | null;
  minimum_score?: string | null;
  valid_from?: string | null;
  valid_to?: string | null;
  created_at: string;
}

export type UpsertRankStandardPayload = Pick<RankStandard, "rank_id" | "discipline_id" | "minimum_place" | "minimum_score" | "valid_from" | "valid_to">;
