"use client";
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { logout as requestLogout } from "../api/auth";
import { getAccessIdentity } from "./jwt";
import { AUTH_CHANGE_EVENT, AUTH_FAILURE_EVENT, tokenStorage } from "./storage";
import type { AuthIdentity } from "./types";
import { useApiReady } from "@/_app/providers/api-provider";

type AuthContextValue = { identity: AuthIdentity | null; isAuthenticated: boolean; isLoading: boolean; logout: () => Promise<void> };
const AuthContext = createContext<AuthContextValue | null>(null);
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const ready = useApiReady(); const [identity, setIdentity] = useState<AuthIdentity | null>(null); const [isLoading, setLoading] = useState(true);
  useEffect(() => {
    if (!ready) return;
    const sync = () => { const token = tokenStorage.get()?.access_token; const next = token ? getAccessIdentity(token) : null; if (next && next.expiresAt > Date.now()) setIdentity(next); else { if (token) tokenStorage.clear(); setIdentity(null); } setLoading(false); };
    sync(); window.addEventListener(AUTH_CHANGE_EVENT, sync); window.addEventListener(AUTH_FAILURE_EVENT, sync);
    return () => { window.removeEventListener(AUTH_CHANGE_EVENT, sync); window.removeEventListener(AUTH_FAILURE_EVENT, sync); };
  }, [ready]);
  const value = useMemo(() => ({ identity, isAuthenticated: Boolean(identity), isLoading, logout: requestLogout }), [identity, isLoading]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
export function useAuth() { const context = useContext(AuthContext); if (!context) throw new Error("useAuth must be used inside AuthProvider"); return context; }
