"use client";

import { useCallback, useMemo, useState } from "react";
import { FileText } from "lucide-react";

import { competitionResultApi } from "@/entities/competition-result";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { ChartLine } from "@/shared/ui/chart";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";

export function ResultsHistory() {
  const load = useCallback(() => competitionResultApi.history(), []);
  const state = useAsyncData(load);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const items = useMemo(() => (state.data ?? []).filter((item) => `${item.competition_name} ${item.discipline_name}`.toLocaleLowerCase("ru").includes(search.trim().toLocaleLowerCase("ru")) && (status === "all" || item.result_status === status)), [search, state.data, status]);
  const chartItems = [...items]
    .filter((item) => item.score != null || item.place != null)
    .sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime());
  const scorePoints = chartItems.filter((item) => item.score != null).map((item) => ({
    label: formatDate(item.start_at),
    value: item.score as number,
  }));
  const placePoints = chartItems.filter((item) => item.place != null).map((item) => ({
    label: formatDate(item.start_at),
    value: item.place as number,
  }));

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar search={search} onSearchChange={setSearch} searchPlaceholder="Найти результат" filterContent={<Select value={status} onValueChange={setStatus}><SelectTrigger><SelectValue placeholder="Статус" /></SelectTrigger><SelectContent><SelectItem value="all">Все статусы</SelectItem><SelectItem value="approved">Утверждённые</SelectItem><SelectItem value="finished">Завершённые</SelectItem></SelectContent></Select>} />
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !items.length ? (
        <DataEmpty title="Результатов пока нет" description="После утверждения результатов соревнований они появятся здесь." />
      ) : null}
      {!state.loading && !state.error && (scorePoints.length >= 2 || placePoints.length >= 2) ? (
        <Card className="mb-4 bg-white/75 p-5 shadow-none">
          <CardHeader className="p-0">
            <CardTitle>Динамика результатов</CardTitle>
            <CardDescription>
              {scorePoints.length >= 2 ? "Баллы по утверждённым соревнованиям." : "Места по утверждённым соревнованиям; меньшее место лучше."}
            </CardDescription>
          </CardHeader>
          <div className="mt-4">
            <ChartLine
              points={scorePoints.length >= 2 ? scorePoints : placePoints}
              label={scorePoints.length >= 2 ? "График изменения баллов по соревнованиям" : "График мест по соревнованиям"}
            />
          </div>
        </Card>
      ) : null}
      <div className="grid gap-3">
        {items.map((item) => (
          <Card key={`${item.competition_id}-${item.event_id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div className="flex min-w-0 gap-3">
                <div>
                  <p className="font-medium">{item.competition_name}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{item.discipline_name} · {formatDate(item.start_at)}</p>
                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    {item.place ? <span className="text-sm font-medium">{item.place} место</span> : null}
                    {item.score != null ? <span className="text-sm text-muted-foreground">{item.score} баллов</span> : null}
                    <StatusBadge status={item.result_status} />
                  </div>
                </div>
              </div>
              {item.protocol_url ? (
                <Button asChild variant="outline" size="sm">
                  <a href={item.protocol_url} target="_blank" rel="noreferrer"><FileText />Протокол</a>
                </Button>
              ) : null}
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
