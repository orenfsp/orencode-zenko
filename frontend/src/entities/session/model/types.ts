export interface Tokens {
  access_token: string;
  refresh_token: string;
}

export interface AuthIdentity {
  userId: number;
  roleId: number | null;
  principalType: string;
  permissions: string[] | null;
  expiresAt: number;
}

export interface SessionMetadata {
  displayName: string;
  email: string;
  roleIds: number[];
}
