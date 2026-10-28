import { WorkspaceScreen } from "@/_pages/workspace";
import { TeamsView } from "@/features/workspace/teams/ui/teams-view";

export default function TeamsPage() {
  return <WorkspaceScreen title="Сборные команды" permissions={["teams.manage"]}><TeamsView /></WorkspaceScreen>;
}
