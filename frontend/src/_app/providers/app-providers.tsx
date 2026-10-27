"use client";
import { AuthProvider } from "@/entities/session/model/auth-provider";
import { NotificationProvider } from "@/shared/lib/notifications";
import { ApiProvider } from "./api-provider";
export function AppProviders({ children }: { children: React.ReactNode }) { return <NotificationProvider><ApiProvider><AuthProvider>{children}</AuthProvider></ApiProvider></NotificationProvider>; }
