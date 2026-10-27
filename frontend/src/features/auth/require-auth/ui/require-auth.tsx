import { AuthGuard } from "@/features/auth/auth-guard";
export function RequireAuth({ children }: { children: React.ReactNode }) { return <AuthGuard redirectTo="/login?next=/workspace">{children}</AuthGuard>; }
