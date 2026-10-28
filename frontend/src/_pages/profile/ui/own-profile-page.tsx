import { ProfileView } from "@/features/profile/view";
import { RequireAuth } from "@/features/auth/require-auth";
import { WorkspaceShell } from "@/widgets/workspace-shell";

export function OwnProfilePage() {
  return <RequireAuth><WorkspaceShell title="Профиль"><ProfileView mode="own" /></WorkspaceShell></RequireAuth>;
}
