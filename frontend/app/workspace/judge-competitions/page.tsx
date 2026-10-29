import { WorkspaceScreen } from "@/_pages/workspace";
import { JudgeAssignments } from "@/features/workspace/judge-assignments";

export default function JudgeCompetitionsPage() {
  return <WorkspaceScreen title="Мои соревнования" roles={["judge"]}><JudgeAssignments /></WorkspaceScreen>;
}
