"use client";
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { logout as requestLogout } from "../api/auth";
import { getAccessIdentity } from "./jwt";
import { AUTH_CHANGE_EVENT, AUTH_FAILURE_EVENT, tokenStorage } from "./storage";
import type { AuthIdentity } from "./types";
import { useApiReady } from "@/_app/providers/api-provider";
import { rbacApi } from "@/entities/rbac";
import { setApiPermissions } from "@/shared/api/client";

type AuthContextValue = { identity: AuthIdentity | null; isAuthenticated: boolean; isLoading: boolean; logout: () => Promise<void> };
const AuthContext = createContext<AuthContextValue | null>(null);
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const ready = useApiReady(); const [identity, setIdentity] = useState<AuthIdentity | null>(null); const [isLoading, setLoading] = useState(true);
  useEffect(() => {
    if (!ready) return;
    let active = true;
    let syncVersion = 0;
    const sync = async () => {
      const version = ++syncVersion;
      const token = tokenStorage.get()?.access_token;
      const next = token ? getAccessIdentity(token) : null;

      if (!next || next.expiresAt <= Date.now()) {
        if (token) tokenStorage.clear();
        if (active && version === syncVersion) {
          setApiPermissions(null);
          setIdentity(null);
          setLoading(false);
        }
        return;
      }

      setLoading(true);
      let resolved = next;
      if (next.roleId !== null) {
        try {
          resolved = { ...next, permissions: await rbacApi.myPermissions() };
        } catch {
          // Keep permissions from the JWT if the permissions endpoint is unavailable.
        }
      }
      if (active && version === syncVersion) {
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
  }, [ready]);
  const value = useMemo(() => ({ identity, isAuthenticated: Boolean(identity), isLoading, logout: requestLogout }), [identity, isLoading]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
export function useAuth() { const context = useContext(AuthContext); if (!context) throw new Error("useAuth must be used inside AuthProvider"); return context; }
