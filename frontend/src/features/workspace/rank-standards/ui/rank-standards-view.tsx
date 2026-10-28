"use client";

import { type FormEvent, useCallback, useState } from "react";
import { rankStandardApi } from "@/entities/rank-standard";
import { rankApi } from "@/entities/rank";
import { referenceDataApi } from "@/entities/reference-data";
import { formatDate } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { DatePicker } from "@/shared/ui/date-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function RankStandardsView() {
  const notify = useNotifications();
  const [pending, setPending] = useState(false);
  const load = useCallback(async () => {
    const [standards, ranks, disciplines] = await Promise.all([rankStandardApi.list(), rankApi.list(), referenceDataApi.disciplines()]);
    return { standards, ranks, disciplines };
  }, []);
  const state = useAsyncData(load);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const optional = (key: string) => String(form.get(key) ?? "").trim() || undefined;
    const minimumPlace = optional("minimum_place");
    const minimumScore = optional("minimum_score");
    if (!minimumPlace && !minimumScore) return notify.error("Укажите место или минимальный балл");
    setPending(true);
    try {
      await rankStandardApi.upsert({
        rank_id: Number(form.get("rank_id")),
        discipline_id: Number(form.get("discipline_id")),
        minimum_place: minimumPlace ? Number(minimumPlace) : undefined,
        minimum_score: minimumScore,
        valid_from: optional("valid_from"),
        valid_to: optional("valid_to"),
      });
      notify.success("Норматив сохранён");
      state.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось сохранить норматив"); }
    finally { setPending(false); }
  }

  return <div className="mx-auto w-full max-w-6xl">
    <Card className="mb-5 bg-white/75 p-5 shadow-none"><form className="grid gap-3" onSubmit={(event) => void submit(event)}>
      <div className="grid gap-3 sm:grid-cols-2">
        <Select name="rank_id" required><SelectTrigger><SelectValue placeholder="Разряд" /></SelectTrigger><SelectContent>{state.data?.ranks.map((rank) => <SelectItem key={rank.id} value={String(rank.id)}>{rank.name}</SelectItem>)}</SelectContent></Select>
        <Select name="discipline_id" required><SelectTrigger><SelectValue placeholder="Дисциплина" /></SelectTrigger><SelectContent>{state.data?.disciplines.map((discipline) => <SelectItem key={discipline.id} value={String(discipline.id)}>{discipline.name}</SelectItem>)}</SelectContent></Select>
        <Input name="minimum_place" type="number" min="1" placeholder="Место не ниже" />
        <Input name="minimum_score" type="number" min="0" step="0.01" placeholder="Минимальный балл" />
        <div><p className="mb-1 text-sm">Действует с</p><DatePicker id="standard-valid-from" name="valid_from" allowFuture /></div>
        <div><p className="mb-1 text-sm">Действует до</p><DatePicker id="standard-valid-to" name="valid_to" allowFuture /></div>
      </div>
      <Button disabled={pending || !state.data} className="w-fit">{pending ? "Сохранение…" : "Сохранить норматив"}</Button>
    </form></Card>
    {state.loading ? <DataLoading rows={4} /> : null}
    {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
    {!state.loading && !state.error && !state.data?.standards.length ? <DataEmpty title="Нормативов пока нет" /> : null}
    <div className="grid gap-2">{state.data?.standards.map((standard) => <Card key={standard.id} className="bg-white/75 p-4 shadow-none"><p className="font-medium">{state.data?.ranks.find((item) => item.id === standard.rank_id)?.name ?? `Разряд #${standard.rank_id}`} · {state.data?.disciplines.find((item) => item.id === standard.discipline_id)?.name ?? `Дисциплина #${standard.discipline_id}`}</p><p className="mt-1 text-sm text-muted-foreground">{standard.minimum_place != null ? `Место до ${standard.minimum_place}` : ""}{standard.minimum_place != null && standard.minimum_score != null ? " · " : ""}{standard.minimum_score != null ? `От ${standard.minimum_score} баллов` : ""}</p><p className="mt-1 text-xs text-muted-foreground">{standard.valid_from ? formatDate(standard.valid_from) : "Без даты начала"} — {standard.valid_to ? formatDate(standard.valid_to) : "бессрочно"}</p></Card>)}</div>
  </div>;
}
