import { WorkspaceScreen } from "@/_pages/workspace";
import { FilesList } from "@/features/workspace/files";

export default function FilesPage() {
  return <WorkspaceScreen title="Файлы" permissions={["files.read"]}><FilesList /></WorkspaceScreen>;
}
