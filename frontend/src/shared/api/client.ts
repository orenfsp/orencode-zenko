import { ApiError, type ApiEnvelope, type ApiErrorPayload } from "@/shared/api/types";

export interface RequestOptions {
  auth?: boolean;
  permission?: string;
  retry?: boolean;
  params?: Record<string, string | number | boolean | undefined>;
  signal?: AbortSignal;
  responseType?: "json" | "blob" | "response";
}

interface ApiClientOptions {
  getAccessToken: () => string | null;
  getPermissions: () => string[] | null;
  refresh: () => Promise<string | null>;
  onAuthFailure: () => void;
}

let options: ApiClientOptions | null = null;
let refreshPromise: Promise<string | null> | null = null;
let runtimePermissions: string[] | null | undefined;

function toUrl(path: string, params?: RequestOptions["params"]) {
  const url = new URL(path, window.location.origin);
  for (const [key, value] of Object.entries(params ?? {})) {
    if (value !== undefined) url.searchParams.set(key, String(value));
  }
  return url.toString();
}

async function parse<T>(response: Response, responseType: RequestOptions["responseType"]): Promise<T> {
  if (response.status === 204) return undefined as T;
  if (response.ok && responseType === "blob") return (await response.blob()) as T;
  if (response.ok && responseType === "response") return response as T;
  const payload = (await response.json().catch(() => ({}))) as T | ApiErrorPayload;
  if (!response.ok) {
    const error = payload as ApiErrorPayload;
    throw new ApiError(error.message || "Не удалось выполнить запрос", response.status, error);
  }
  return payload as T;
}

async function request<T>(
  path: string,
  init: RequestInit = {},
  requestOptions: RequestOptions = {},
): Promise<T> {
  if (!options) throw new Error("API-клиент не инициализирован");
  const requiresAuth = requestOptions.auth !== false;
  const headers = new Headers(init.headers);

  if (requiresAuth) {
    const token = options.getAccessToken();
    if (!token) {
      options.onAuthFailure();
      throw new ApiError("Необходима авторизация", 401);
    }
    const permissions = options.getPermissions();
    if (
      requestOptions.permission &&
      permissions !== null &&
      !permissions.includes(requestOptions.permission)
    ) {
      throw new ApiError("Недостаточно прав для этого действия", 403);
    }
    headers.set("Authorization", `Bearer ${token}`);
  }

  if (init.body && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }
  headers.set("Accept", "application/json");

  let response: Response;
  try {
    response = await fetch(toUrl(path, requestOptions.params), {
      ...init,
      headers,
      signal: requestOptions.signal,
    });
  } catch {
    throw new ApiError("Ошибка сети", 0);
  }

  if (response.status === 401 && requiresAuth && !requestOptions.retry) {
    refreshPromise ??= options.refresh().finally(() => {
      refreshPromise = null;
    });
    let token: string | null;
    try {
      token = await refreshPromise;
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 401) options.onAuthFailure();
      throw cause;
    }
    if (token) return request<T>(path, init, { ...requestOptions, retry: true });
    options.onAuthFailure();
    throw new ApiError("Необходима авторизация", 401);
  }

  if (response.status === 401 && requiresAuth && requestOptions.retry) {
    options.onAuthFailure();
  }

  return parse<T>(response, requestOptions.responseType);
}

export const api = {
  get<T>(path: string, options?: RequestOptions) {
    return request<T>(path, { method: "GET" }, options);
  },
  post<T>(path: string, body?: unknown, options?: RequestOptions) {
    return request<T>(path, { method: "POST", body: body instanceof FormData ? body : JSON.stringify(body) }, options);
  },
  patch<T>(path: string, body?: unknown, options?: RequestOptions) {
    return request<T>(path, { method: "PATCH", body: body instanceof FormData ? body : JSON.stringify(body) }, options);
  },
  put<T>(path: string, body?: unknown, options?: RequestOptions) {
    return request<T>(path, { method: "PUT", body: body instanceof FormData ? body : JSON.stringify(body) }, options);
  },
  delete<T>(path: string, options?: RequestOptions) {
    return request<T>(path, { method: "DELETE" }, options);
  },
};

export function initApiClient(nextOptions: ApiClientOptions) {
  options = nextOptions;
}

export function resetApiClient() {
  options = null;
  refreshPromise = null;
  runtimePermissions = undefined;
}

export function setApiPermissions(permissions: string[] | null) {
  runtimePermissions = permissions;
}

export function getApiPermissions() {
  return runtimePermissions;
}

export { ApiError } from "./types";
export type { ApiEnvelope };
