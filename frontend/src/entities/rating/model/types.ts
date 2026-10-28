export interface RatingEntry {
  position: number;
  user_id: number;
  last_name: string;
  first_name: string;
  middle_name?: string | null;
  municipality_name?: string | null;
  total_points: number;
  result_count: number;
  wins: number;
}

export interface RatingPage {
  items: RatingEntry[];
  limit: number;
  offset: number;
}

export interface RatingQuery {
  year?: number;
  discipline_id?: number;
  age_group_id?: number;
  municipality_id?: number;
  limit?: number;
  offset?: number;
}

export interface RatingSelectionMember {
  selection_position: number;
  user_id: number;
  last_name: string;
  first_name: string;
  middle_name?: string | null;
}

export interface RatingSelection {
  id: number;
  season_year: number;
  discipline: { id: number; name: string };
  age_group: { id: number; name: string };
  status: "draft" | "locked";
  captain_user_id: number | null;
  coach_user_id: number | null;
  members: RatingSelectionMember[];
  created_by_user_id: number;
  locked_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface RatingSelectionQuery {
  year: number;
  discipline_id: number;
  age_group_id: number;
}

export interface SaveRatingSelectionPayload {
  season_year: number;
  discipline_id: number;
  age_group_id: number;
  member_ids: number[];
  captain_user_id: number | null;
  coach_user_id: number | null;
}
