import { WorkspaceScreen } from "@/_pages/workspace";
import { AdministrationView } from "@/features/workspace/administration";

export default function AdministrationReviewsPage() {
  return <WorkspaceScreen title="Проверки" permissions={["anti_doping.review"]}><AdministrationView section="reviews" /></WorkspaceScreen>;
}
