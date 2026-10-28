import { WorkspaceScreen } from "@/_pages/workspace";
import { AdministrationView } from "@/features/workspace/administration";

export default function AdministrationReviewsPage() {
  return <WorkspaceScreen title="Проверка сертификатов" permissions={["anti_doping.review"]}><AdministrationView section="reviews" /></WorkspaceScreen>;
}
