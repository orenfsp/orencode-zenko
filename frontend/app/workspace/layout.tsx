import { RequireAuth } from "@/features/auth/require-auth";
import { WorkspaceShell } from "@/widgets/workspace-shell";

export default function WorkspaceLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <RequireAuth>
      <WorkspaceShell>{children}</WorkspaceShell>
    </RequireAuth>
  );
}
