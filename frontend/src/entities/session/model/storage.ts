import type { Tokens } from "./types";

const accessKey = "zenko_access_token";
const refreshKey = "zenko_refresh_token";
export const AUTH_FAILURE_EVENT = "zenko:auth-failure";
export const AUTH_CHANGE_EVENT = "zenko:auth-change";

function notify(event: string) {
  if (typeof window !== "undefined") window.dispatchEvent(new Event(event));
}

export const tokenStorage = {
  get(): Tokens | null {
    if (typeof window === "undefined") return null;
    const access_token = localStorage.getItem(accessKey);
    const refresh_token = localStorage.getItem(refreshKey);
    return access_token && refresh_token ? { access_token, refresh_token } : null;
  },
  set(tokens: Tokens) {
    localStorage.setItem(accessKey, tokens.access_token);
    localStorage.setItem(refreshKey, tokens.refresh_token);
    notify(AUTH_CHANGE_EVENT);
  },
  clear() {
    localStorage.removeItem(accessKey);
    localStorage.removeItem(refreshKey);
    notify(AUTH_CHANGE_EVENT);
  },
};
