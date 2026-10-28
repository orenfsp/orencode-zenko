"use client";

import { useCallback, useMemo, useState } from "react";

import { rankApplicationApi } from "@/entities/rank-application";
import { formatDate } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { ChartLine } from "@/shared/ui/chart";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

export function RankHistory() {
  const load = useCallback(() => rankApplicationApi.history(), []);
  const state = useAsyncData(load);
  const [search, setSearch] = useState("");
  const items = useMemo(() => (state.data ?? []).filter((rank) => `${rank.rank_name} ${rank.discipline_name}`.toLocaleLowerCase("ru").includes(search.trim().toLocaleLowerCase("ru"))), [search, state.data]);
  const awardedPoints = [...items]
    .filter((rank) => rank.awarded_at)
    .sort((a, b) => new Date(a.awarded_at as string).getTime() - new Date(b.awarded_at as string).getTime())
    .map((rank, index) => ({ label: formatDate(rank.awarded_at), value: index + 1 }));

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar search={search} onSearchChange={setSearch} searchPlaceholder="Найти разряд" filterContent={<p className="text-sm text-muted-foreground">История отображается по дате присвоения.</p>} />
      {state.loading ? <DataLoading rows={3} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !items.length ? (
        <DataEmpty title="Разрядов пока нет" description="Подтверждённые присвоения появятся на этой странице." />
      ) : null}
      {!state.loading && !state.error && awardedPoints.length >= 2 ? (
        <Card className="mb-4 bg-white/75 p-5 shadow-none">
          <CardHeader className="p-0">
            <CardTitle>Динамика присвоений</CardTitle>
            <CardDescription>Накопительное количество подтверждённых разрядов по датам.</CardDescription>
          </CardHeader>
          <div className="mt-4"><ChartLine points={awardedPoints} label="График накопительного количества присвоенных разрядов" /></div>
        </Card>
      ) : null}
      <div className="grid gap-3">
        {items.map((rank) => (
          <Card key={rank.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex gap-3">
              <div>
                <p className="font-medium">{rank.rank_name}</p>
                <p className="mt-1 text-sm text-muted-foreground">{rank.discipline_name}</p>
                <div className="mt-3 flex flex-wrap gap-x-5 gap-y-1 text-xs text-muted-foreground">
                  <span>Присвоен: {formatDate(rank.awarded_at)}</span>
                  <span>Действует до: {formatDate(rank.expires_at)}</span>
                  {rank.order_number ? <span>Приказ № {rank.order_number}</span> : null}
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
