import { WorkspaceScreen } from "@/_pages/workspace";
import { CompetitionDetails } from "@/features/workspace/competition-details";

export default async function CompetitionPage({
  params,
}: {
  params: Promise<{ competitionId: string }>;
}) {
  const { competitionId } = await params;
  return (
    <WorkspaceScreen title="Соревнование">
      <CompetitionDetails competitionId={Number(competitionId)} />
    </WorkspaceScreen>
  );
}
