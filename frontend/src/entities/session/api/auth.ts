import { api, ApiError, type ApiEnvelope } from "@/shared/api/client";
import { isAccessTokenValid } from "../model/jwt";
import { tokenStorage } from "../model/storage";
import type { Tokens } from "../model/types";

export { ApiError };
export async function login(email: string, password: string) {
  return (
    await api.post<ApiEnvelope<Tokens>>(
      "/api/v1/auth/login",
      { email, password },
      { auth: false },
    )
  ).data;
}
export async function register(body: FormData) {
  return (
    await api.post<ApiEnvelope<Tokens>>("/api/v1/auth/register", body, {
      auth: false,
    })
  ).data;
}
export async function chooseRole(roleId: number, tokens: Tokens) {
  return (
    await api.patch<ApiEnvelope<Tokens & { role_id: number }>>(
      "/api/v1/profile",
      { role_id: roleId, refresh_token: tokens.refresh_token },
    )
  ).data;
}
export async function chooseRoles(roleIds: number[], tokens: Tokens) {
  return (
    await api.patch<
      ApiEnvelope<Tokens & { role_id: number; role_ids: number[] }>
    >("/api/v1/profile", {
      role_ids: roleIds,
      refresh_token: tokens.refresh_token,
    })
  ).data;
}
export async function refreshSession() {
  const tokens = tokenStorage.get();
  if (!tokens) return null;
  const result = await api.post<ApiEnvelope<Tokens>>(
    "/api/v1/auth/refresh",
    { refresh_token: tokens.refresh_token },
    { auth: false },
  );
  tokenStorage.set(result.data);
  return result.data;
}
export async function logout() {
  const tokens = tokenStorage.get();
  try {
    if (tokens)
      await api.post<void>(
        "/api/v1/auth/logout",
        { refresh_token: tokens.refresh_token },
        { auth: false },
      );
  } finally {
    tokenStorage.clear();
  }
}
export function storeTokens(tokens: Tokens) {
  tokenStorage.set(tokens);
}
export function readTokens() {
  return tokenStorage.get();
}
export { isAccessTokenValid };
export function accessTokenHasRole(accessToken: string) {
  const payload = accessToken.split(".")[1];
  if (!payload) return false;
  try {
    return (
      typeof (
        JSON.parse(
          atob(
            payload
              .replace(/-/g, "+")
              .replace(/_/g, "/")
              .padEnd(Math.ceil(payload.length / 4) * 4, "="),
          ),
        ) as { role_id?: unknown }
      ).role_id === "number"
    );
  } catch {
    return false;
  }
}
export function splitFullName(value: string) {
  const parts = value.trim().split(/\s+/).filter(Boolean);
  return {
    firstName: parts[0] ?? "",
    lastName: parts[1] ?? "",
    middleName: parts.slice(2).join(" "),
  };
}
export function isMinor(birthDate: string) {
  const birth = new Date(`${birthDate}T00:00:00`);
  const today = new Date();
  let age = today.getFullYear() - birth.getFullYear();
  if (
    today.getMonth() < birth.getMonth() ||
    (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())
  )
    age -= 1;
  return age < 18;
}
