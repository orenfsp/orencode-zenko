"use client";

import { useCallback } from "react";
import { CalendarRange } from "lucide-react";

import { calendarPlanApi } from "@/entities/calendar-plan";
import { SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function CalendarPlansList() {
  const load = useCallback(() => calendarPlanApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading title="Календарные планы" description="Официальные планы и периоды проведения соревнований." />
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? <DataEmpty title="Календарных планов пока нет" /> : null}
      <div className="grid gap-3">
        {state.data?.map((plan) => (
          <Card key={plan.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><CalendarRange className="size-5" /></span>
                <div>
                  <p className="font-medium">{plan.name}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{plan.year} год</p>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
