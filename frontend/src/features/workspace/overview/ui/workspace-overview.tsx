"use client";

import * as React from "react";
import Link from "next/link";
import { CalendarDays } from "lucide-react";

import { competitionApi, type Competition } from "@/entities/competition";
import { competitionApplicationApi, type CompetitionApplication } from "@/entities/competition-application";
import { competitionJudgeApi, type JudgeAssignment } from "@/entities/competition-judge";
import { dashboardApi, type DashboardData } from "@/entities/dashboard";
import { deadlineApi, type DeadlineItem } from "@/entities/deadline";
import { judgeCategoryApi, type CurrentJudgeCategory, type JudgeCategoryApplication } from "@/entities/judge-category";
import { profileApi } from "@/entities/profile";
import { userCoachApi, type UserCoachRelation } from "@/entities/user-coach";
import { useAuth } from "@/entities/session";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

type PersonalOverview = {
  kind: "personal";
  role: string;
  deadlines: DeadlineItem[];
  competitions: Competition[];
  applications: CompetitionApplication[];
  relations: UserCoachRelation[];
  judgeAssignments: JudgeAssignment[];
  judgeApplications: JudgeCategoryApplication[];
  currentCategory: CurrentJudgeCategory | null;
};

type OverviewData =
  | { kind: "federation"; dashboard: DashboardData }
  | PersonalOverview;

const metricCards = [
  { key: "athletes", label: "Спортсмены" },
  { key: "competitions", label: "Соревнования" },
  { key: "participants", label: "Участники" },
  { key: "active_ranks", label: "Активные разряды" },
] as const;

function Metric({ label, value }: { label: string; value: number | string }) {
  return (
    <Card className="bg-white/75 p-4 shadow-none sm:p-5">
      <p className="text-2xl font-medium tabular-nums">{value}</p>
      <p className="mt-1 text-sm text-muted-foreground">{label}</p>
    </Card>
  );
}

function SectionCard({ title, description, children }: { title: string; description?: string; children: React.ReactNode }) {
  return (
    <Card className="bg-white/75 p-5 shadow-none">
      <CardHeader className="p-0">
        <CardTitle>{title}</CardTitle>
        {description ? <CardDescription>{description}</CardDescription> : null}
      </CardHeader>
      <div className="mt-4">{children}</div>
    </Card>
  );
}

function QueueRow({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center justify-between rounded-xl bg-muted/60 px-3 py-3 text-sm">
      <span>{label}</span>
      <span className="font-medium tabular-nums">{value}</span>
    </div>
  );
}

function UpcomingCompetitions({ competitions }: { competitions: Competition[] }) {
  const upcoming = competitions
    .filter((competition) => competition.status !== "cancelled")
    .sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime())
    .slice(0, 5);

  return (
    <div className="grid gap-3">
      {upcoming.map((competition) => (
        <Link key={competition.id} href={`/workspace/competitions/${competition.id}`} className="flex items-center justify-between gap-3 rounded-xl bg-muted/60 p-3 transition-colors hover:bg-muted">
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">{competition.name}</p>
            <p className="text-xs text-muted-foreground">{formatDate(competition.start_at)} · {competition.discipline.name}</p>
          </div>
          <StatusBadge status={competition.status} />
        </Link>
      ))}
      {!upcoming.length ? <p className="text-sm text-muted-foreground">Ближайших соревнований пока нет.</p> : null}
    </div>
  );
}

function DeadlineList({ deadlines }: { deadlines: DeadlineItem[] }) {
  return (
    <div className="grid gap-3">
      {deadlines.slice(0, 5).map((deadline) => (
        <div key={`${deadline.type}-${deadline.id}`} className="flex items-start justify-between gap-3 rounded-xl bg-muted/60 p-3">
          <div className="flex min-w-0 gap-3">
            <CalendarDays className="mt-0.5 size-4 shrink-0 text-primary" />
            <div className="min-w-0">
              <p className="text-sm font-medium">{deadline.title}</p>
              <p className="mt-1 text-xs text-muted-foreground">До {formatDate(deadline.due_at, true)}</p>
            </div>
          </div>
          <StatusBadge status={deadline.status} />
        </div>
      ))}
      {!deadlines.length ? <p className="text-sm text-muted-foreground">Ближайших сроков нет.</p> : null}
    </div>
  );
}

function PersonalDashboard({ data }: { data: PersonalOverview }) {
  const isCoach = data.role === "coach";
  const isJudge = data.role === "judge";
  const confirmedRelations = data.relations.filter((relation) => relation.status === "confirmed");
  const pendingApplications = data.applications.filter((application) => application.status === "pending");
  const pendingJudgeApplications = data.judgeApplications.filter((application) => application.status === "pending");

  if (isJudge) {
    return (
      <div className="mx-auto grid w-full max-w-6xl gap-5">
        <div className="grid gap-3 sm:grid-cols-3">
          <Metric label="Назначения" value={data.judgeAssignments.length} />
          <Metric label="Заявки на категорию" value={pendingJudgeApplications.length} />
          <Metric label="Текущая категория" value={data.currentCategory?.category_name ?? "Не назначена"} />
        </div>
        <div className="grid gap-5 lg:grid-cols-2">
          <SectionCard title="Ближайшие назначения" description="Соревнования, где вы участвуете в судейской коллегии.">
            <div className="grid gap-3">
              {data.judgeAssignments.slice(0, 5).map((assignment) => (
                <Link key={assignment.id} href={`/workspace/competitions/${assignment.competition.id}`} className="flex items-center justify-between gap-3 rounded-xl bg-muted/60 p-3 transition-colors hover:bg-muted">
                  <div className="min-w-0"><p className="truncate text-sm font-medium">{assignment.competition.name}</p><p className="text-xs text-muted-foreground">{assignment.position.name} · {formatDate(assignment.competition.start_at)}</p></div>
                  <StatusBadge status={assignment.competition.status} />
                </Link>
              ))}
              {!data.judgeAssignments.length ? <p className="text-sm text-muted-foreground">Назначений пока нет.</p> : null}
            </div>
          </SectionCard>
          <SectionCard title="Ваши сроки" description="Задачи, которые требуют внимания."><DeadlineList deadlines={data.deadlines} /></SectionCard>
        </div>
        <div className="flex flex-wrap gap-2"><Button asChild><Link href="/workspace/judge-competitions">Мои соревнования</Link></Button><Button asChild variant="outline"><Link href="/workspace/applications">Мои заявки</Link></Button></div>
      </div>
    );
  }

  return (
    <div className="mx-auto grid w-full max-w-6xl gap-5">
      <div className="grid gap-3 sm:grid-cols-3">
        <Metric label={isCoach ? "Подтверждённые спортсмены" : "Мои тренеры"} value={confirmedRelations.length} />
        <Metric label="Заявки на рассмотрении" value={pendingApplications.length} />
        <Metric label="Ближайшие сроки" value={data.deadlines.length} />
      </div>
      <div className="grid gap-5 lg:grid-cols-2">
        <SectionCard title={isCoach ? "Мои спортсмены" : "Мои заявки"} description={isCoach ? "Связи и заявки спортсменов в одном месте." : "Текущие заявки на участие в соревнованиях."}>
          {isCoach ? <div className="grid gap-3"><QueueRow label="Подтверждённые связи" value={confirmedRelations.length} /><QueueRow label="Ожидают подтверждения" value={data.relations.filter((relation) => relation.status === "pending").length} /><Button asChild variant="outline" className="mt-1 w-full"><Link href="/workspace/coaches">Открыть спортсменов</Link></Button></div> : <div className="grid gap-3"><QueueRow label="На рассмотрении" value={pendingApplications.length} /><QueueRow label="Всего заявок" value={data.applications.length} /><Button asChild variant="outline" className="mt-1 w-full"><Link href="/workspace/applications">Открыть заявки</Link></Button></div>}
        </SectionCard>
        <SectionCard title="Ближайшие соревнования" description="Мероприятия, которые могут быть важны для вас."><UpcomingCompetitions competitions={data.competitions} /></SectionCard>
      </div>
      <SectionCard title="Ваши сроки" description="Не пропустите важные даты."><DeadlineList deadlines={data.deadlines} /></SectionCard>
      <div className="flex flex-wrap gap-2"><Button asChild><Link href="/workspace/competitions">Найти соревнование</Link></Button>{isCoach ? <Button asChild variant="outline"><Link href="/workspace/applications">Заявки спортсменов</Link></Button> : <Button asChild variant="outline"><Link href="/workspace/profile">Открыть профиль</Link></Button>}</div>
    </div>
  );
}

export function WorkspaceOverview() {
  const { identity } = useAuth();
  const canReadDashboard = identity?.permissions?.includes("dashboard.read") ?? false;

  const load = React.useCallback(async (): Promise<OverviewData> => {
    if (canReadDashboard) return { kind: "federation", dashboard: await dashboardApi.get() };

    const profile = await profileApi.me();
    const role = profile.roles.find((item) => item.id === identity?.roleId)?.code ?? "participant";
    const deadlines = await deadlineApi.list();

    if (role === "judge") {
      const [judgeAssignments, judgeApplications, currentCategory] = await Promise.all([
        competitionJudgeApi.profileAssignments({ limit: 100 }),
        judgeCategoryApi.mine(),
        judgeCategoryApi.current(),
      ]);
      return { kind: "personal", role, deadlines, competitions: [], applications: [], relations: [], judgeAssignments, judgeApplications, currentCategory };
    }

    const [competitions, applications, relations] = await Promise.all([
      competitionApi.list({ limit: 100 }),
      competitionApplicationApi.mine(),
      role === "coach" ? userCoachApi.byCoach({ limit: 100 }) : userCoachApi.byParticipant({ limit: 100 }),
    ]);
    return { kind: "personal", role, deadlines, competitions, applications, relations, judgeAssignments: [], judgeApplications: [], currentCategory: null };
  }, [canReadDashboard, identity?.roleId]);

  const state = useAsyncData(load);
  if (state.loading) return <DataLoading rows={4} />;
  if (state.error) return <DataError message={state.error} retry={state.reload} />;
  if (!state.data) return <DataEmpty title="Данные пока недоступны" description="Попробуйте обновить страницу позже." />;
  if (state.data.kind === "personal") return <PersonalDashboard data={state.data} />;

  const dashboard = state.data.dashboard;
  return (
    <div className="mx-auto grid w-full max-w-6xl gap-5">
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">{metricCards.map(({ key, label }) => <Metric key={key} label={label} value={dashboard.metrics[key]} />)}</div>
      <div className="grid gap-5 lg:grid-cols-2">
        <SectionCard title="Очередь на рассмотрение" description="Заявки, которые требуют решения федерации."><div className="grid gap-3"><QueueRow label="Заявки на соревнования" value={dashboard.pending_competition_applications.length} /><QueueRow label="Заявки на разряды" value={dashboard.pending_rank_applications.length} /><QueueRow label="Судейские категории" value={dashboard.pending_judge_category_applications.length} /><Button asChild variant="outline" className="mt-1 w-full"><Link href="/workspace/applications">Открыть заявки</Link></Button></div></SectionCard>
        <SectionCard title="Ближайшие соревнования" description="Опубликованные мероприятия из календаря."><div className="grid gap-3">{dashboard.upcoming_competitions.slice(0, 5).map((competition) => <Link key={competition.id} href={`/workspace/competitions/${competition.id}`} className="flex items-center justify-between gap-3 rounded-xl bg-muted/60 p-3 transition-colors hover:bg-muted"><div className="min-w-0"><p className="truncate text-sm font-medium">{competition.name}</p><p className="text-xs text-muted-foreground">{formatDate(competition.start_at)} · {competition.discipline_name}</p></div><StatusBadge status={competition.status} /></Link>)}{!dashboard.upcoming_competitions.length ? <p className="text-sm text-muted-foreground">Соревнований пока нет.</p> : null}</div></SectionCard>
      </div>
      <SectionCard title="Контрольные сроки" description="Ближайшие сроки и документы, требующие внимания."><DeadlineList deadlines={dashboard.deadlines.map((item, index) => ({ ...item, id: item.competition_id ?? index, status: "pending" }))} /></SectionCard>
      <div className="flex flex-wrap gap-2"><Button asChild><Link href="/workspace/registry">Открыть реестр</Link></Button><Button asChild variant="outline"><Link href="/workspace/administration/audit">Журнал аудита</Link></Button></div>
    </div>
  );
}
