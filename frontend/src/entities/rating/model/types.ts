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
  municipality_id?: number;
  limit?: number;
  offset?: number;
}
