"use client";
import { AuthProvider } from "@/entities/session/model/auth-provider";
import { NotificationProvider } from "@/shared/lib/notifications";
import { ApiProvider } from "./api-provider";
import { ScrollMaskController } from "@/shared/ui/scroll-mask";
import { ThemeProvider } from "./theme-provider";

export function AppProviders({ children }: { children: React.ReactNode }) {
  return <ThemeProvider><NotificationProvider><ApiProvider><AuthProvider><ScrollMaskController />{children}</AuthProvider></ApiProvider></NotificationProvider></ThemeProvider>;
}
