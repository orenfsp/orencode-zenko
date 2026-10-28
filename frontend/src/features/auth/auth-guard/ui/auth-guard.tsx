"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/entities/session";
import { cn } from "@/shared/lib/utils";
import { Spinner } from "@/shared/ui/spinner";
import Link from "next/link";
import { Button } from "@/shared/ui/button";
interface AuthGuardProps { children: React.ReactNode; requireAuth?: boolean; redirectTo?: string; fallback?: React.ReactNode; className?: string; permissions?: readonly string[]; permissionMode?: "any" | "all"; }
export function AuthGuard({ children, requireAuth = true, redirectTo = "/login", fallback, className, permissions, permissionMode = "any" }: AuthGuardProps) {
  const { identity, isAuthenticated, isLoading, permissionsError, sessionError, retry } = useAuth(); const router = useRouter(); const knownPermissions = identity?.permissions;
  const hasAccess = !permissions?.length || (Array.isArray(knownPermissions) && (permissionMode === "all" ? permissions.every((permission) => knownPermissions.includes(permission)) : permissions.some((permission) => knownPermissions.includes(permission))));
  useEffect(() => {
    if (isLoading || sessionError) return;
    if (requireAuth && !isAuthenticated) router.replace(redirectTo);
    else if (!requireAuth && isAuthenticated) router.replace(identity?.roleId == null ? "/choose-role" : redirectTo);
  }, [identity?.roleId, isAuthenticated, isLoading, redirectTo, requireAuth, router, sessionError]);
  if (isLoading) return fallback ?? <main className="flex min-h-svh items-center justify-center"><Spinner className="size-5 text-primary" /></main>;
  if (sessionError) return <AccessState title="Не удалось проверить сессию" description="Проверьте подключение и попробуйте ещё раз. Ваши данные входа сохранены." retry={retry} />;
  if ((requireAuth ? !isAuthenticated : isAuthenticated)) return null;
  if (permissions?.length && permissionsError) return <AccessState title="Не удалось проверить права" description="Сервис проверки прав временно недоступен. Попробуйте повторить запрос." retry={retry} />;
  if (!hasAccess) return <AccessState title="Нет доступа к разделу" description="Для этого действия у активной роли недостаточно прав. Вернитесь в рабочее пространство или обратитесь к администратору." />;
  return <div className={cn("flex min-h-0 flex-1 flex-col", className)}>{children}</div>;
}

function AccessState({ title, description, retry }: { title: string; description: string; retry?: () => void }) {
  return <main className="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center gap-4 px-6 text-center" role="alert">
    <h1 className="text-xl font-semibold">{title}</h1>
    <p className="text-sm text-muted-foreground">{description}</p>
    <div className="flex flex-wrap justify-center gap-2">{retry ? <Button onClick={retry}>Повторить</Button> : null}<Button asChild variant="outline"><Link href="/workspace">К обзору</Link></Button></div>
  </main>;
}
