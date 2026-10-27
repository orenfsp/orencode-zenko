import type { AuthIdentity } from "./types";

interface AccessClaims {
  user_id?: number;
  role_id?: number;
  principal_type?: string;
  permissions?: string[];
  exp?: number;
}

export function getAccessIdentity(token: string): AuthIdentity | null {
  try {
    const payload = token.split(".")[1];
    if (!payload) return null;
    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const claims = JSON.parse(atob(normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "="))) as AccessClaims;
    if (typeof claims.user_id !== "number" || typeof claims.exp !== "number") return null;
    return {
      userId: claims.user_id,
      roleId: typeof claims.role_id === "number" ? claims.role_id : null,
      principalType: claims.principal_type ?? "user",
      permissions: Array.isArray(claims.permissions) ? claims.permissions : null,
      expiresAt: claims.exp * 1000,
    };
  } catch {
    return null;
  }
}

export function isAccessTokenValid(token: string) {
  const identity = getAccessIdentity(token);
  return Boolean(identity && identity.expiresAt > Date.now());
}
