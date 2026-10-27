import { RequireAuth } from "@/features/auth/require-auth";
import { WorkspaceShell } from "@/widgets/workspace-shell";
export function WorkspacePage() { return <RequireAuth><WorkspaceShell /></RequireAuth>; }
