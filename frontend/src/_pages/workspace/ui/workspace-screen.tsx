import { AuthGuard } from "@/features/auth/auth-guard";
import { RequireAuth } from "@/features/auth/require-auth";
import { WorkspaceShell } from "@/widgets/workspace-shell";

export function WorkspaceScreen({
  title,
  children,
  permissions,
}: {
  title: string;
  children: React.ReactNode;
  permissions?: readonly string[];
}) {
  return (
    <RequireAuth>
      <AuthGuard permissions={permissions} redirectTo="/workspace">
        <WorkspaceShell title={title}>{children}</WorkspaceShell>
      </AuthGuard>
    </RequireAuth>
  );
}
