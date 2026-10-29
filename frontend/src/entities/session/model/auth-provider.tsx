"use client";
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { ApiError, logout as requestLogout, refreshSession } from "../api/auth";
import { getAccessIdentity } from "./jwt";
import { AUTH_CHANGE_EVENT, AUTH_FAILURE_EVENT, tokenStorage } from "./storage";
import type { AuthIdentity } from "./types";
import { useApiReady } from "@/_app/providers/api-provider";
import { rbacApi } from "@/entities/rbac";
import { setApiPermissions } from "@/shared/api/client";

type AuthContextValue = { identity: AuthIdentity | null; isAuthenticated: boolean; isLoading: boolean; permissionsError: boolean; sessionError: boolean; retry: () => void; logout: () => Promise<void> };
const AuthContext = createContext<AuthContextValue | null>(null);
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const ready = useApiReady(); const [identity, setIdentity] = useState<AuthIdentity | null>(null); const [isLoading, setLoading] = useState(true);
  const [permissionsError, setPermissionsError] = useState(false);
  const [sessionError, setSessionError] = useState(false);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    if (!ready) return;
    let active = true;
    let syncVersion = 0;
    const sync = async () => {
      const version = ++syncVersion;
      const storedTokens = tokenStorage.get();
      let token = storedTokens?.access_token ?? null;
      let next = token ? getAccessIdentity(token) : null;

      // Access-токен может истечь между переходами по страницам. Сначала
      // пробуем бесшумно обновить сессию по refresh-токену и только затем
      // считаем пользователя разлогиненным.
      if (storedTokens && (!next || next.expiresAt <= Date.now())) {
        let refreshed = null;
        try {
          refreshed = await refreshSession();
        } catch (cause) {
          if (!(cause instanceof ApiError) || cause.status !== 401) {
            if (active && version === syncVersion) {
              setSessionError(true);
              setLoading(false);
            }
            return;
          }
        }
        if (refreshed) {
          token = refreshed.access_token;
          next = getAccessIdentity(token);
        }
      }

      if (!next || next.expiresAt <= Date.now()) {
        if (token) tokenStorage.clear();
        if (active && version === syncVersion) {
          setSessionError(false);
          setPermissionsError(false);
          setApiPermissions(null);
          setIdentity(null);
          setLoading(false);
        }
        return;
      }

      setLoading(true);
      setSessionError(false);
      let resolved = next;
      let failedPermissions = false;
      if (next.roleId !== null) {
        try {
          resolved = {
            ...next,
            permissions: await rbacApi.permissionsForActiveRole(),
          };
        } catch {
          resolved = { ...next, permissions: null };
          failedPermissions = true;
        }
      }
      if (active && version === syncVersion) {
        setPermissionsError(failedPermissions);
        setApiPermissions(resolved.permissions);
        setIdentity(resolved);
        setLoading(false);
      }
    };
    const handleAuthChange = () => void sync();
    void sync();
    window.addEventListener(AUTH_CHANGE_EVENT, handleAuthChange);
    window.addEventListener(AUTH_FAILURE_EVENT, handleAuthChange);
    return () => {
      active = false;
      window.removeEventListener(AUTH_CHANGE_EVENT, handleAuthChange);
      window.removeEventListener(AUTH_FAILURE_EVENT, handleAuthChange);
    };
  }, [ready, revision]);
  const retry = useCallback(() => setRevision((value) => value + 1), []);
  const value = useMemo(() => ({ identity, isAuthenticated: Boolean(identity), isLoading, permissionsError, sessionError, retry, logout: requestLogout }), [identity, isLoading, permissionsError, sessionError, retry]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
export function useAuth() { const context = useContext(AuthContext); if (!context) throw new Error("useAuth must be used inside AuthProvider"); return context; }
