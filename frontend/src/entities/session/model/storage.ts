import type { SessionMetadata, Tokens } from "./types";

const accessKey = "zenko_access_token";
const refreshKey = "zenko_refresh_token";
const metadataKey = "zenko_session_metadata";
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
    localStorage.removeItem(metadataKey);
    notify(AUTH_CHANGE_EVENT);
  },
};

const emptyMetadata: SessionMetadata = {
  displayName: "",
  email: "",
  roleIds: [],
};

export const sessionMetadataStorage = {
  get(): SessionMetadata {
    if (typeof window === "undefined") return emptyMetadata;
    try {
      const value = JSON.parse(
        localStorage.getItem(metadataKey) ?? "{}",
      ) as Partial<SessionMetadata>;
      return {
        displayName:
          typeof value.displayName === "string" ? value.displayName : "",
        email: typeof value.email === "string" ? value.email : "",
        roleIds: Array.isArray(value.roleIds)
          ? value.roleIds.filter(
              (roleId): roleId is number =>
                typeof roleId === "number" && Number.isInteger(roleId),
            )
          : [],
      };
    } catch {
      return emptyMetadata;
    }
  },
  set(value: Partial<SessionMetadata>) {
    if (typeof window === "undefined") return;
    const current = this.get();
    localStorage.setItem(
      metadataKey,
      JSON.stringify({ ...current, ...value }),
    );
    notify(AUTH_CHANGE_EVENT);
  },
};
