export interface UpdateProfilePayload { tg_username?: string; role_id?: number; role_ids?: number[]; refresh_token?: string }
export interface UpdateProfileResult {
  tg_username?: string;
  role_id?: number;
  role_ids?: number[];
  access_token?: string;
  refresh_token?: string;
}
