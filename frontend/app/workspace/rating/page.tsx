import { WorkspaceScreen } from "@/_pages/workspace";
import { RatingView } from "@/features/workspace/rating/ui/rating-view";

export default function RatingPage() {
  return <WorkspaceScreen title="Рейтинг спортсменов" permissions={["rating.read"]}><RatingView /></WorkspaceScreen>;
}
