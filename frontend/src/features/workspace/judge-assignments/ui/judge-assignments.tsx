"use client";

import { useCallback } from "react";

import { competitionJudgeApi } from "@/entities/competition-judge";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function JudgeAssignments() {
  const load = useCallback(
    () => competitionJudgeApi.profileAssignments({ limit: 100 }),
    [],
  );
  const state = useAsyncData(load);

  return (
    <div className="mx-auto w-full max-w-6xl">
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? (
        <DataEmpty title="Назначений пока нет" description="После назначения федерацией соревнование появится здесь." />
      ) : null}
      <div className="grid gap-3">
        {state.data?.map((assignment) => (
          <Card key={assignment.id} variant="glass" className="p-5">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <div>
                  <p className="font-medium">{assignment.competition.name}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{assignment.position.name}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{formatDate(assignment.competition.start_at)} — {formatDate(assignment.competition.end_at)}</p>
                </div>
              </div>
              <StatusBadge status={assignment.competition.status} />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
