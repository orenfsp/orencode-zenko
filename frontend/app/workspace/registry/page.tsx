import { WorkspaceScreen } from "@/_pages/workspace";
import { RegistryView } from "@/features/workspace/registry";

export default function RegistryPage() {
  return <WorkspaceScreen title="Реестр" permissions={["registry.read"]}><RegistryView /></WorkspaceScreen>;
}
