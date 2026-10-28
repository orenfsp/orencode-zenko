import { WorkspaceScreen } from "@/_pages/workspace";
import { DeadlinesList } from "@/features/workspace/deadlines";

export default function DeadlinesPage() {
  return <WorkspaceScreen title="Сроки"><DeadlinesList /></WorkspaceScreen>;
}
