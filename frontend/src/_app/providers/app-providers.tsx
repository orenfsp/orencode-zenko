"use client";
import { AuthProvider } from "@/entities/session/model/auth-provider";
import { NotificationProvider } from "@/shared/lib/notifications";
import { ApiProvider } from "./api-provider";
import { ScrollMaskController } from "@/shared/ui/scroll-mask";
export function AppProviders({ children }: { children: React.ReactNode }) { return <NotificationProvider><ApiProvider><AuthProvider><ScrollMaskController />{children}</AuthProvider></ApiProvider></NotificationProvider>; }
