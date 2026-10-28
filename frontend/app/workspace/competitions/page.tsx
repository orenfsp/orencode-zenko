import { WorkspaceScreen } from "@/_pages/workspace";
import { CompetitionsList } from "@/features/workspace/competitions";

export default function CompetitionsPage() {
  return <WorkspaceScreen title="Соревнования"><CompetitionsList /></WorkspaceScreen>;
}
