import { WorkspaceScreen } from "@/_pages/workspace";
import { RankStandardsView } from "@/features/workspace/rank-standards/ui/rank-standards-view";

export default function RankStandardsPage() {
  return <WorkspaceScreen title="Нормативы разрядов" permissions={["rank_applications.review"]}><RankStandardsView /></WorkspaceScreen>;
}
