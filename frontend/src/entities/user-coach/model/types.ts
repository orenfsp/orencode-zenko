export type UserCoachStatus = "pending" | "confirmed" | "ended";

export interface UserCoachRelation {
  id: number;
  participant_user_id: number;
  coach_user_id: number;
  discipline_id: number;
  status: UserCoachStatus;
  requested_at: string;
  confirmed_at?: string;
  ended_at?: string;
}

export interface CreateUserCoachPayload {
  participant_user_id: number;
  discipline_id: number;
}

export interface UserCoachListQuery {
  limit?: number;
  offset?: number;
}
