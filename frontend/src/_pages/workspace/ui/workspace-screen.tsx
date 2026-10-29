import { AuthGuard } from "@/features/auth/auth-guard";

export function WorkspaceScreen({
  children,
  permissions,
  roles,
}: {
  title?: string;
  children: React.ReactNode;
  permissions?: readonly string[];
  roles?: readonly string[];
}) {
  return <AuthGuard permissions={permissions} roles={roles} redirectTo="/workspace">{children}</AuthGuard>;
}
