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

export interface CoachParticipantDiscipline {
  id: number;
  name: string;
}

export interface CoachParticipantRelation {
  discipline_id: number;
  status: UserCoachStatus;
}

export interface CoachParticipant {
  id: number;
  first_name: string;
  last_name: string;
  middle_name: string | null;
  organization: string | null;
  municipality: string | null;
  disciplines: CoachParticipantDiscipline[];
  coach_relations: CoachParticipantRelation[];
}

export interface CoachParticipantPage {
  items: CoachParticipant[];
  total: number;
  limit: number;
  offset: number;
}

export interface CoachParticipantQuery extends UserCoachListQuery {
  search?: string;
  discipline_id?: number;
}
