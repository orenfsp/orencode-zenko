export interface UpdateProfilePayload { tg_username?: string; role_id?: number; refresh_token?: string }
export interface UpdateProfileResult { access_token: string; refresh_token: string; role_id?: number }
