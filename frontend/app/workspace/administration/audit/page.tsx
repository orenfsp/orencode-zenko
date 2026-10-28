import { WorkspaceScreen } from "@/_pages/workspace";
import { AdministrationView } from "@/features/workspace/administration";

export default function AdministrationAuditPage() {
  return <WorkspaceScreen title="Журнал аудита" permissions={["roles.read"]}><AdministrationView section="audit" /></WorkspaceScreen>;
}
