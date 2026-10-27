"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { initApiClient, resetApiClient } from "@/shared/api/client";
import { AUTH_FAILURE_EVENT, getAccessIdentity, refreshSession, tokenStorage } from "@/entities/session";

const ApiContext = createContext(false);

export function ApiProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false);
  useEffect(() => {
    initApiClient({
      getAccessToken: () => tokenStorage.get()?.access_token ?? null,
      getPermissions: () => { const token = tokenStorage.get()?.access_token; return token ? getAccessIdentity(token)?.permissions ?? null : null; },
      refresh: async () => (await refreshSession())?.access_token ?? null,
      onAuthFailure: () => { tokenStorage.clear(); window.dispatchEvent(new Event(AUTH_FAILURE_EVENT)); },
    });
    queueMicrotask(() => setReady(true));
    return resetApiClient;
  }, []);
  return <ApiContext.Provider value={ready}>{children}</ApiContext.Provider>;
}
export const useApiReady = () => useContext(ApiContext);
