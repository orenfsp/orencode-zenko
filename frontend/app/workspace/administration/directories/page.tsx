import { WorkspaceScreen } from "@/_pages/workspace";
import { AdministrationView } from "@/features/workspace/administration";

export default function AdministrationDirectoriesPage() {
  return <WorkspaceScreen title="Справочники" permissions={["municipalities.create", "rank_applications.review"]}><AdministrationView section="directories" /></WorkspaceScreen>;
}
