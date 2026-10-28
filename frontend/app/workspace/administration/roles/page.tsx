import { WorkspaceScreen } from "@/_pages/workspace";
import { AdministrationView } from "@/features/workspace/administration";

export default function AdministrationRolesPage() {
  return <WorkspaceScreen title="Роли и права" permissions={["roles.read", "roles.create"]}><AdministrationView section="roles" /></WorkspaceScreen>;
}
