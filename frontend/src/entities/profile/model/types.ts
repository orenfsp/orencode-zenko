export interface ProfileReference { id: number; name: string }
export interface ProfileRole { id: number; code: string; name: string }
export interface ProfileRank { id: number; rank_id: number; rank_name: string; discipline_id: number; discipline_name: string; awarded_at?: string; expires_at?: string; order_number?: string }
export interface ProfileJudgeCategory { id: number; category_id: number; category_name: string; assigned_at: string; expires_at?: string; order_number?: string }
export interface PublicProfile {
  id: number; last_name: string; first_name: string; middle_name?: string; age: number;
  organization: ProfileReference | null; municipality: ProfileReference | null;
  roles: ProfileRole[]; disciplines: ProfileReference[]; ranks: ProfileRank[];
  judge_category: ProfileJudgeCategory | null;
}
export interface OwnProfile extends PublicProfile { birth_date: string; email: string; tg_username: string | null; access_status: boolean }
export interface UpdateProfilePayload { tg_username?: string; role_id?: number; role_ids?: number[]; refresh_token?: string }
export interface UpdateProfileResult { tg_username?: string; role_id?: number; role_ids?: number[]; access_token?: string; refresh_token?: string }
