"use client";
import * as React from "react";
import { CalendarDays, ClipboardList, Medal, Trophy, Users } from "lucide-react";
import Link from "next/link";
import { LoaderCircle } from "lucide-react";

import { dashboardApi, type DashboardData } from "@/entities/dashboard";
import { deadlineApi, type DeadlineItem } from "@/entities/deadline";
import { profileApi } from "@/entities/profile";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { DataEmpty, DataLoading } from "@/shared/ui/data-state";

type OverviewData =
  | { kind: "federation"; dashboard: DashboardData }
  | { kind: "personal"; deadlines: DeadlineItem[] };

const metricCards = [
  { key: "athletes", label: "Спортсмены", icon: Users },
  { key: "competitions", label: "Соревнования", icon: Trophy },
  { key: "participants", label: "Участники", icon: ClipboardList },
  { key: "active_ranks", label: "Активные разряды", icon: Medal },
] as const;


export function WorkspaceOverview() {
  const { identity } = useAuth();

  const canReadDashboard =
    identity?.permissions?.includes("dashboard.read") ?? false;

  const [data, setData] = React.useState<OverviewData | null>(null);

  const [loading, setLoading] = React.useState(true);

  const load = React.useCallback(async (): Promise<OverviewData> => {
    if (canReadDashboard) {
      return {
        kind: "federation",
        dashboard: await dashboardApi.get(),
      };
    }

    const profile = await profileApi.me();

    const activeRole = profile.roles.find(
      (role) => role.id === identity?.roleId,
    )?.code;

    if (
      activeRole === "federation" ||
      activeRole === "super_admin"
    ) {
      return {
        kind: "federation",
        dashboard: await dashboardApi.get(),
      };
    }

    return {
      kind: "personal",
      deadlines: await deadlineApi.list(),
    };
  }, [canReadDashboard, identity?.roleId]);

  React.useEffect(() => {
    let cancelled = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;

    const fetchData = async () => {
      try {
        const result = await load();

        if (cancelled) return;

        setData(result);
        setLoading(false);
      } catch (error) {
        if (cancelled) return;

        // Ошибку не показываем пользователю.
        // Повторяем запрос через 3 секунды.
        console.error(
          "[WorkspaceOverview] Ошибка загрузки. Повтор через 3 секунды:",
          error,
        );

        retryTimer = setTimeout(fetchData, 3000);
      }
    };

    setLoading(true);
    setData(null);

    void fetchData();

    return () => {
      cancelled = true;

      if (retryTimer !== undefined) {
        clearTimeout(retryTimer);
      }
    };
  }, [load]);

  

if (loading || !data) {
  return (
    <div className="flex min-h-[60vh] w-full h-full items-center justify-center">
      <LoaderCircle className="size-8 animate-spin text-primary" />
    </div>
  );
}

  if (data.kind === "personal") {
    return (
      <div className="mx-auto w-full max-w-5xl">
        <SectionHeading
          title="Ближайшие события"
          description="Сроки и действия, сформированные для вашей активной роли."
          action={
            <Button asChild variant="outline" size="sm">
              <Link href="/workspace/competitions">
                Все соревнования
              </Link>
            </Button>
          }
        />

        {data.deadlines.length ? (
          <div className="grid gap-3">
            {data.deadlines.slice(0, 8).map((deadline) => (
              <Card
                key={`${deadline.type}-${deadline.id}`}
                className="bg-white/75 p-5 shadow-none"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex min-w-0 gap-3">
                    <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary">
                      <CalendarDays className="size-5" />
                    </span>

                    <div className="min-w-0">
                      <p className="font-medium">
                        {deadline.title}
                      </p>

                      <p className="mt-1 text-sm text-muted-foreground">
                        До {formatDate(deadline.due_at, true)}
                      </p>
                    </div>
                  </div>

                  <StatusBadge status={deadline.status} />
                </div>
              </Card>
            ))}
          </div>
        ) : (
          <DataEmpty
            title="Ближайших сроков нет"
            description="Новые события появятся здесь автоматически."
          />
        )}
      </div>
    );
  }

  const dashboard = data.dashboard;

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading
        title="Сводка федерации"
        description="Ключевые показатели и ближайшие рабочие задачи."
      />

      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {metricCards.map(({ key, label, icon: Icon }) => (
          <Card
            key={key}
            className="bg-white/75 p-5 shadow-none"
          >
            <span className="mb-4 flex size-10 items-center justify-center rounded-xl bg-[#eeebff] text-primary">
              <Icon className="size-5" />
            </span>

            <p className="text-2xl font-medium">
              {dashboard.metrics[key]}
            </p>

            <p className="mt-1 text-sm text-muted-foreground">
              {label}
            </p>
          </Card>
        ))}
      </div>

      <div className="mt-5 grid gap-5 lg:grid-cols-2">
        <Card className="bg-white/75 p-5 shadow-none">
          <CardHeader>
            <CardTitle>Ближайшие соревнования</CardTitle>
            <CardDescription>
              Опубликованные мероприятия из календаря.
            </CardDescription>
          </CardHeader>

          <div className="mt-4 grid gap-3">
            {dashboard.upcoming_competitions
              .slice(0, 5)
              .map((competition) => (
                <div
                  key={competition.id}
                  className="flex items-center justify-between gap-3 rounded-xl bg-muted/60 p-3"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">
                      {competition.name}
                    </p>

                    <p className="text-xs text-muted-foreground">
                      {formatDate(competition.start_at)} ·{" "}
                      {competition.discipline_name}
                    </p>
                  </div>

                  <StatusBadge status={competition.status} />
                </div>
              ))}

            {!dashboard.upcoming_competitions.length && (
              <p className="text-sm text-muted-foreground">
                Соревнований пока нет.
              </p>
            )}
          </div>
        </Card>

        <Card className="bg-white/75 p-5 shadow-none">
          <CardHeader>
            <CardTitle>Очереди на рассмотрение</CardTitle>
            <CardDescription>
              Заявки, которые требуют решения федерации.
            </CardDescription>
          </CardHeader>

          <div className="mt-4 grid gap-3">
            <QueueRow
              label="Заявки на соревнования"
              value={
                dashboard.pending_competition_applications.length
              }
            />

            <QueueRow
              label="Заявки на разряды"
              value={dashboard.pending_rank_applications.length}
            />

            <QueueRow
              label="Судейские категории"
              value={
                dashboard.pending_judge_category_applications.length
              }
            />

            <Button
              asChild
              variant="outline"
              className="mt-1 w-full"
            >
              <Link href="/workspace/applications">
                Открыть заявки
              </Link>
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
}


function QueueRow({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center justify-between rounded-xl bg-muted/60 px-3 py-3 text-sm">
      <span>{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  );
}
