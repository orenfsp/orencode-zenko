"use client";

import { FormEvent, useCallback, useState } from "react";
import { CalendarDays, LoaderCircle, Pencil, Search } from "lucide-react";

import { calendarPlanApi } from "@/entities/calendar-plan";
import type { CalendarPlan, CalendarPlanCompetition } from "@/entities/calendar-plan";
import { useAuth } from "@/entities/session";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useNotifications } from "@/shared/lib/notifications";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

type SelectedPlan = {
  plan: CalendarPlan;
  competitions: CalendarPlanCompetition[];
};

export function CalendarPlansList() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const canCreate = identity?.permissions?.includes("calendar_plans.create") ?? false;
  const canUpdate = identity?.permissions?.includes("calendar_plans.update") ?? false;
  const [pending, setPending] = useState(false);
  const [search, setSearch] = useState("");
  const [yearFilter, setYearFilter] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<CalendarPlan | null>(null);
  const [selected, setSelected] = useState<SelectedPlan | null>(null);
  const [detailsOpen, setDetailsOpen] = useState(false);

  const load = useCallback(() => calendarPlanApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);

  const plans = (state.data ?? []).filter((plan) => {
    const query = search.trim().toLocaleLowerCase("ru");
    const year = yearFilter.trim();
    const matchesSearch = !query || `${plan.name} ${plan.year}`.toLocaleLowerCase("ru").includes(query);
    const matchesYear = !year || String(plan.year) === year;
    return matchesSearch && matchesYear;
  });

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPending(true);
    try {
      await calendarPlanApi.create({ name: String(form.get("name") ?? "").trim(), year: Number(form.get("year")) });
      notify.success("Календарный план создан");
      setCreateOpen(false);
      event.currentTarget.reset();
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось создать план");
    } finally {
      setPending(false);
    }
  }

  async function update(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editing) return;
    const form = new FormData(event.currentTarget);
    setPending(true);
    try {
      await calendarPlanApi.update(editing.id, { name: String(form.get("name") ?? "").trim(), year: Number(form.get("year")) });
      notify.success("План обновлён");
      setEditing(null);
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось обновить план");
    } finally {
      setPending(false);
    }
  }

  async function openPlan(plan: CalendarPlan) {
    setPending(true);
    try {
      const competitions = await calendarPlanApi.competitions(plan.id, { limit: 100 });
      setSelected({ plan, competitions });
      setDetailsOpen(true);
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить соревнования плана");
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Найти план по названию или году"
        filterContent={
          <div className="grid gap-2">
            <label htmlFor="calendar-plan-year-filter" className="text-sm font-medium">Год плана</label>
            <Input id="calendar-plan-year-filter" type="number" min="1" placeholder="Например, 2026" value={yearFilter} onChange={(event) => setYearFilter(event.target.value)} />
          </div>
        }
        createLabel="Создать план"
        onCreate={canCreate ? () => setCreateOpen(true) : undefined}
      />

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Новый календарный план</DialogTitle>
            <DialogDescription>Создайте план, чтобы затем привязывать к нему соревнования.</DialogDescription>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={create}>
            <div className="grid gap-2"><label htmlFor="calendar-plan-name" className="text-sm font-medium">Название</label><Input id="calendar-plan-name" name="name" placeholder="Календарный план на 2026 год" required /></div>
            <div className="grid gap-2"><label htmlFor="calendar-plan-year" className="text-sm font-medium">Год</label><Input id="calendar-plan-year" name="year" type="number" min="1" max="32767" placeholder="2026" required /></div>
            <DialogFooter><Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Создать план</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(editing)} onOpenChange={(open) => { if (!open) setEditing(null); }}>
        <DialogContent>
          <DialogHeader><DialogTitle>Изменить календарный план</DialogTitle><DialogDescription>Изменения применятся только к названию и году плана.</DialogDescription></DialogHeader>
          {editing ? <form key={editing.id} className="grid gap-4" onSubmit={update}>
            <div className="grid gap-2"><label htmlFor="edit-calendar-plan-name" className="text-sm font-medium">Название</label><Input id="edit-calendar-plan-name" name="name" defaultValue={editing.name} required /></div>
            <div className="grid gap-2"><label htmlFor="edit-calendar-plan-year" className="text-sm font-medium">Год</label><Input id="edit-calendar-plan-year" name="year" type="number" min="1" max="32767" defaultValue={editing.year} required /></div>
            <DialogFooter><Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Сохранить изменения</Button></DialogFooter>
          </form> : null}
        </DialogContent>
      </Dialog>

      <Dialog open={detailsOpen} onOpenChange={setDetailsOpen}>
        <DialogContent className="max-h-[min(720px,calc(100vh-2rem))] overflow-y-auto scroll-mask-y sm:max-w-2xl">
          <DialogHeader><DialogTitle>{selected?.plan.name}</DialogTitle><DialogDescription>Соревнования, включённые в календарный план за {selected?.plan.year} год.</DialogDescription></DialogHeader>
          {selected ? selected.competitions.length ? <div className="grid gap-2">
            {selected.competitions.map((competition) => <div key={competition.id} className="rounded-xl border bg-muted/30 p-3"><div className="flex flex-wrap items-start justify-between gap-3"><div className="min-w-0"><p className="font-medium">{competition.name}</p><p className="mt-1 text-xs text-muted-foreground">{competition.discipline.name} · {formatDate(competition.start_at)} — {formatDate(competition.end_at)}{competition.calendar_plan_number ? ` · № ${competition.calendar_plan_number}` : ""}</p></div><StatusBadge status={competition.status} /></div></div>)}
          </div> : <div className="rounded-xl border border-dashed p-6 text-center"><CalendarDays className="mx-auto size-5 text-muted-foreground" /><p className="mt-2 text-sm font-medium">В плане пока нет соревнований</p><p className="mt-1 text-sm text-muted-foreground">Привяжите план при создании или редактировании черновика соревнования.</p></div> : null}
        </DialogContent>
      </Dialog>

      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !plans.length ? <DataEmpty title={search || yearFilter ? "Планы не найдены" : "Календарных планов пока нет"} description={search || yearFilter ? "Измените запрос или сбросьте фильтр по году." : canCreate ? "Создайте первый план кнопкой «Создать план», чтобы связать с ним соревнования." : "Планы появятся после того, как администратор добавит их."} /> : null}

      <div className="grid gap-3">
        {plans.map((plan) => <Card key={plan.id} className="bg-white/75 p-4 shadow-none sm:p-5"><div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div className="min-w-0"><div className="flex items-center gap-2"><CalendarDays className="size-4 shrink-0 text-muted-foreground" /><h2 className="truncate font-medium">{plan.name}</h2></div><p className="mt-1 pl-6 text-sm text-muted-foreground">{plan.year} год · план № {plan.id}</p></div><div className="flex flex-wrap gap-2 sm:shrink-0"><Button type="button" variant="outline" onClick={() => void openPlan(plan)} disabled={pending}><Search />Открыть соревнования</Button>{canUpdate ? <Button type="button" variant="ghost" size="icon" aria-label={`Изменить план «${plan.name}»`} onClick={() => setEditing(plan)} disabled={pending}><Pencil /></Button> : null}</div></div></Card>)}
      </div>
    </div>
  );
}
