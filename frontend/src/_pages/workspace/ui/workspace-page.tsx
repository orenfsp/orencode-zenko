import { RequireAuth } from "@/features/auth/require-auth";
import { WorkspaceOverview } from "@/features/workspace/overview";
import { WorkspaceShell } from "@/widgets/workspace-shell";
export function WorkspacePage() {
  return (
    <RequireAuth>
      <WorkspaceShell>
        <WorkspaceOverview />
      </WorkspaceShell>
    </RequireAuth>
  );
}
