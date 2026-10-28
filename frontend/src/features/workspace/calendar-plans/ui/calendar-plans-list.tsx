"use client";

import { FormEvent, useCallback, useState } from "react";
import { CalendarRange, LoaderCircle, Plus, Save } from "lucide-react";

import { calendarPlanApi } from "@/entities/calendar-plan";
import type { CalendarPlan, CalendarPlanCompetition } from "@/entities/calendar-plan";
import { useAuth } from "@/entities/session";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";

export function CalendarPlansList() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const canCreate = identity?.permissions?.includes("calendar_plans.create") ?? false;
  const canUpdate = identity?.permissions?.includes("calendar_plans.update") ?? false;
  const [pending, setPending] = useState(false);
  const [selected, setSelected] = useState<{ plan: CalendarPlan; competitions: CalendarPlanCompetition[] } | null>(null);
  const load = useCallback(() => calendarPlanApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPending(true);
    try {
      await calendarPlanApi.create({ name: String(form.get("name") ?? ""), year: Number(form.get("year")) });
      notify.success("Календарный план создан");
      event.currentTarget.reset();
      state.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось создать план"); }
    finally { setPending(false); }
  }

  async function update(event: FormEvent<HTMLFormElement>, planId: number) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPending(true);
    try {
      await calendarPlanApi.update(planId, { name: String(form.get("name") ?? ""), year: Number(form.get("year")) });
      notify.success("План обновлён");
      state.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось обновить план"); }
    finally { setPending(false); }
  }

  async function openPlan(planId: number) {
    setPending(true);
    try {
      const [plan, competitions] = await Promise.all([calendarPlanApi.get(planId), calendarPlanApi.competitions(planId, { limit: 100 })]);
      setSelected({ plan, competitions });
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить план"); }
    finally { setPending(false); }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      {canCreate ? <Card className="mb-4 bg-white/75 p-4 shadow-none"><form className="grid gap-2 sm:grid-cols-[1fr_140px_auto]" onSubmit={create}><Input name="name" placeholder="Название плана" required /><Input name="year" type="number" min="2000" max="2200" placeholder="Год" required /><Button disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : <Plus />}Создать</Button></form></Card> : null}
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? <DataEmpty title="Календарных планов пока нет" /> : null}
      <div className="grid gap-3">
        {state.data?.map((plan) => (
          <Card key={plan.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><CalendarRange className="size-5" /></span>
                {canUpdate ? <form className="grid gap-2 sm:grid-cols-[1fr_110px_auto]" onSubmit={(event) => void update(event, plan.id)}><Input name="name" defaultValue={plan.name} required /><Input name="year" type="number" defaultValue={plan.year} required /><Button type="submit" size="sm" variant="outline" disabled={pending}><Save />Сохранить</Button></form> : <div><p className="font-medium">{plan.name}</p><p className="mt-1 text-sm text-muted-foreground">{plan.year} год</p></div>}
              </div>
              <Button size="sm" variant="outline" disabled={pending} onClick={() => void openPlan(plan.id)}>Соревнования</Button>
            </div>
          </Card>
        ))}
      </div>
      {selected ? <Card className="mt-5 bg-white/75 p-5 shadow-none"><SectionHeading title={selected.plan.name} description={`Соревнования календарного плана за ${selected.plan.year} год.`} /><div className="grid gap-2">{selected.competitions.map((competition) => <div key={competition.id} className="flex items-center justify-between gap-3 rounded-xl bg-muted/60 p-3"><div><p className="font-medium">{competition.name}</p><p className="mt-1 text-xs text-muted-foreground">{competition.discipline.name} · {formatDate(competition.start_at)} — {formatDate(competition.end_at)}{competition.calendar_plan_number ? ` · № ${competition.calendar_plan_number}` : ""}</p></div><StatusBadge status={competition.status} /></div>)}{!selected.competitions.length ? <p className="text-sm text-muted-foreground">Соревнования ещё не добавлены.</p> : null}</div></Card> : null}
    </div>
  );
}
