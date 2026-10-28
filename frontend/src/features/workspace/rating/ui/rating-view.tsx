"use client";

import { useCallback, useState } from "react";
import Link from "next/link";
import { ratingApi } from "@/entities/rating";
import { referenceDataApi } from "@/entities/reference-data";
import { municipalityApi } from "@/entities/municipality";
import { SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function RatingView() {
  const [year, setYear] = useState(new Date().getFullYear());
  const [disciplineId, setDisciplineId] = useState(0);
  const [municipalityId, setMunicipalityId] = useState(0);
  const [offset, setOffset] = useState(0);
  const load = useCallback(() => ratingApi.list({ year, discipline_id: disciplineId || undefined, municipality_id: municipalityId || undefined, limit: 50, offset }), [year, disciplineId, municipalityId, offset]);
  const state = useAsyncData(load);
  const references = useAsyncData(useCallback(async () => {
    const [disciplines, municipalities] = await Promise.all([referenceDataApi.disciplines(), municipalityApi.list({ limit: 100 })]);
    return { disciplines, municipalities };
  }, []));

  return <div className="mx-auto w-full max-w-6xl">
    <SectionHeading title="Рейтинг спортсменов" description="Баллы за завершённые соревнования выбранного года." />
    <div className="mb-5 grid gap-3 sm:grid-cols-3">
      <Input aria-label="Год рейтинга" type="number" min={2000} max={new Date().getFullYear() + 1} value={year} onChange={(event) => { setYear(Number(event.target.value)); setOffset(0); }} />
      <Select value={String(disciplineId)} onValueChange={(value) => { setDisciplineId(Number(value)); setOffset(0); }}><SelectTrigger><SelectValue placeholder="Все дисциплины" /></SelectTrigger><SelectContent><SelectItem value="0">Все дисциплины</SelectItem>{references.data?.disciplines.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
      <Select value={String(municipalityId)} onValueChange={(value) => { setMunicipalityId(Number(value)); setOffset(0); }}><SelectTrigger><SelectValue placeholder="Все муниципалитеты" /></SelectTrigger><SelectContent><SelectItem value="0">Все муниципалитеты</SelectItem>{references.data?.municipalities.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
    </div>
    {state.loading ? <DataLoading rows={5} /> : null}
    {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
    {!state.loading && !state.error && !state.data?.items.length ? <DataEmpty title="Результатов нет" description="Попробуйте другой год или фильтр." /> : null}
    <div className="grid gap-2">{state.data?.items.map((entry) => <Card key={entry.user_id} className="flex items-center justify-between gap-3 bg-white/75 p-4 shadow-none"><div><p className="font-medium">{entry.position}. <Link className="hover:underline" href={`/profile/${entry.user_id}`}>{entry.last_name} {entry.first_name} {entry.middle_name ?? ""}</Link></p><p className="text-sm text-muted-foreground">{entry.municipality_name ?? "Муниципалитет не указан"} · {entry.result_count} результатов · {entry.wins} побед</p></div><p className="font-medium">{entry.total_points} баллов</p></Card>)}</div>
    <div className="mt-4 flex gap-2"><Button variant="outline" disabled={offset === 0 || state.loading} onClick={() => setOffset(Math.max(0, offset - 50))}>Назад</Button><Button variant="outline" disabled={state.loading || (state.data?.items.length ?? 0) < 50} onClick={() => setOffset(offset + 50)}>Далее</Button></div>
  </div>;
}
