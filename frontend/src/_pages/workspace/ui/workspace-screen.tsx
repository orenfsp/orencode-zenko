import { AuthGuard } from "@/features/auth/auth-guard";

export function WorkspaceScreen({
  children,
  permissions,
}: {
  title?: string;
  children: React.ReactNode;
  permissions?: readonly string[];
}) {
  return <AuthGuard permissions={permissions} redirectTo="/workspace">{children}</AuthGuard>;
}
