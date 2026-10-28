"use client";

import { useCallback } from "react";
import { Medal } from "lucide-react";

import { rankApplicationApi } from "@/entities/rank-application";
import { formatDate } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function RankHistory() {
  const load = useCallback(() => rankApplicationApi.history(), []);
  const state = useAsyncData(load);

  return (
    <div className="mx-auto w-full max-w-5xl">
      {state.loading ? <DataLoading rows={3} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? (
        <DataEmpty title="Разрядов пока нет" description="Подтверждённые присвоения появятся на этой странице." />
      ) : null}
      <div className="grid gap-3">
        {state.data?.map((rank) => (
          <Card key={rank.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex gap-3">
              <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Medal className="size-5" /></span>
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
