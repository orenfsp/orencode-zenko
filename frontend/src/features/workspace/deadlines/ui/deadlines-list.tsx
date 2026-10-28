"use client";

import { useCallback } from "react";
import { CalendarClock } from "lucide-react";

import { deadlineApi } from "@/entities/deadline";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function DeadlinesList() {
  const load = useCallback(() => deadlineApi.list(), []);
  const state = useAsyncData(load);

  return (
    <div className="mx-auto w-full max-w-5xl">
      <SectionHeading title="Сроки" description="Персональные сроки и обязательные действия для активной роли." />
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? (
        <DataEmpty title="Активных сроков нет" description="Новые сроки появятся автоматически." />
      ) : null}
      <div className="grid gap-3">
        {state.data?.map((deadline) => (
          <Card key={`${deadline.type}-${deadline.id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><CalendarClock className="size-5" /></span>
                <div>
                  <p className="font-medium">{deadline.title}</p>
                  <p className="mt-1 text-sm text-muted-foreground">До {formatDate(deadline.due_at, true)}</p>
                  {deadline.user_name ? <p className="mt-1 text-xs text-muted-foreground">{deadline.user_name}</p> : null}
                </div>
              </div>
              <StatusBadge status={deadline.status} />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
