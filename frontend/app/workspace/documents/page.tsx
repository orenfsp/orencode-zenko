import { WorkspaceScreen } from "@/_pages/workspace";
import { AntiDopingDocuments } from "@/features/workspace/documents";

export default function DocumentsPage() {
  return <WorkspaceScreen title="Антидопинг" roles={["participant"]}><AntiDopingDocuments /></WorkspaceScreen>;
}
