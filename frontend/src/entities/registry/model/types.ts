export interface RegistryReference {
  id: number;
  name: string;
}

export interface RegistryRole {
  id: number;
  code: string;
  name: string;
}

export interface RegistryRank {
  id: number;
  rank_id: number;
  rank_name: string;
  discipline_id: number;
  discipline_name: string;
  awarded_at: string | null;
  expires_at: string | null;
  order_number: string | null;
}

export interface RegistryJudgeCategory {
  id: number;
  category_id: number;
  category_name: string;
  assigned_at: string;
  expires_at: string | null;
  order_number: string | null;
}

export interface RegistryUser {
  id: number;
  last_name: string;
  first_name: string;
  middle_name: string | null;
  birth: string;
  created_at: string;
  organization: RegistryReference | null;
  municipality: RegistryReference | null;
  roles: RegistryRole[];
  disciplines: RegistryReference[];
  ranks: RegistryRank[];
  judge_category: RegistryJudgeCategory | null;
  coach_relations: RegistryCoachRelation[];
  judge_assignments: RegistryJudgeAssignment[];
}

export interface RegistryDynamicsPoint {
  month: string;
  new_users: number;
  total_users: number;
}

export interface RegistryCoachRelation {
  id: number;
  relation_type: "coach" | "participant";
  related_user_id: number;
  related_user_name: string;
  discipline?: RegistryReference | null;
  status: string;
  requested_at: string;
  confirmed_at?: string | null;
  ended_at?: string | null;
}

export interface RegistryJudgeAssignment {
  id: number;
  competition_id: number;
  competition_name: string;
  competition_status: string;
  position_id: number;
  position_name: string;
  start_at: string;
  end_at: string;
  appointed_at: string;
}

export interface RegistryOrganization {
  id: number;
  name: string;
  organization_type: string;
  municipality: RegistryReference | null;
}

export interface RegistryPage<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface RegistryUserQuery {
  search?: string;
  discipline_id?: number;
  organization_id?: number;
  rank_id?: number;
  age_group_id?: number;
  municipality_id?: number;
  role?: string;
  limit?: number;
  offset?: number;
}

export interface RegistryOrganizationQuery {
  search?: string;
  municipality_id?: number;
  limit?: number;
  offset?: number;
}
