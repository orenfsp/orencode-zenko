export interface TeamMember {
  user_id: number;
  last_name: string;
  first_name: string;
  middle_name?: string | null;
}

export interface Team {
  id: number;
  name: string;
  members: TeamMember[];
  coaches: TeamMember[];
  created_at: string;
}

export interface CreateTeamPayload {
  name: string;
  member_ids: number[];
  coach_ids?: number[];
}
