"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/entities/session";
import { cn } from "@/shared/lib/utils";
import { Spinner } from "@/shared/ui/spinner";
interface AuthGuardProps { children: React.ReactNode; requireAuth?: boolean; redirectTo?: string; fallback?: React.ReactNode; className?: string; permissions?: readonly string[]; permissionMode?: "any" | "all"; }
export function AuthGuard({ children, requireAuth = true, redirectTo = "/login", fallback, className, permissions, permissionMode = "any" }: AuthGuardProps) {
  const { identity, isAuthenticated, isLoading } = useAuth(); const router = useRouter(); const knownPermissions = identity?.permissions;
  const hasAccess = !permissions?.length || !Array.isArray(knownPermissions) || (permissionMode === "all" ? permissions.every((permission) => knownPermissions.includes(permission)) : permissions.some((permission) => knownPermissions.includes(permission)));
  useEffect(() => { if (!isLoading && ((requireAuth && !isAuthenticated) || (!requireAuth && isAuthenticated) || !hasAccess)) router.replace(redirectTo); }, [hasAccess, isAuthenticated, isLoading, redirectTo, requireAuth, router]);
  if (isLoading) return fallback ?? <main className="flex min-h-svh items-center justify-center"><Spinner className="size-5 text-primary" /></main>;
  if ((requireAuth ? !isAuthenticated : isAuthenticated) || !hasAccess) return null;
  return <div className={cn("flex min-h-0 flex-1 flex-col", className)}>{children}</div>;
}
