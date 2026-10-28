import { WorkspaceScreen } from "@/_pages/workspace";
import { CalendarPlansList } from "@/features/workspace/calendar-plans";

export default function CalendarPlansPage() {
  return <WorkspaceScreen title="Календарные планы" permissions={["calendar_plans.read"]}><CalendarPlansList /></WorkspaceScreen>;
}
