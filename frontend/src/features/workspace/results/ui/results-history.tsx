"use client";

import { useCallback } from "react";
import { Award, FileText } from "lucide-react";

import { competitionResultApi } from "@/entities/competition-result";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function ResultsHistory() {
  const load = useCallback(() => competitionResultApi.history(), []);
  const state = useAsyncData(load);

  return (
    <div className="mx-auto w-full max-w-5xl">
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? (
        <DataEmpty title="Результатов пока нет" description="После утверждения результатов соревнований они появятся здесь." />
      ) : null}
      <div className="grid gap-3">
        {state.data?.map((item) => (
          <Card key={`${item.competition_id}-${item.event_id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div className="flex min-w-0 gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Award className="size-5" /></span>
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
