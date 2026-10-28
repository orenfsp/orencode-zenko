import { WorkspaceScreen } from "@/_pages/workspace";
import { EventManagement } from "@/features/workspace/event-management";

export default function EventsPage() {
  return <WorkspaceScreen title="Зачёты"><EventManagement /></WorkspaceScreen>;
}
