import { WorkspaceScreen } from "@/_pages/workspace";
import { CoachRelations } from "@/features/workspace/coaches";

export default function CoachesPage() {
  return <WorkspaceScreen title="Тренеры и спортсмены" roles={["participant", "coach"]}><CoachRelations /></WorkspaceScreen>;
}
