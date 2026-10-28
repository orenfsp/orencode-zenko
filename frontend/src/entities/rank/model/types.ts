export interface Rank {
  id: number;
  name: string;
  ttl: number;
  created_at: string;
  updated_at: string;
}

export interface CreateRankPayload {
  name: string;
  ttl: number;
}
