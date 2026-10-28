"use client";

import { useCallback, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ratingApi } from "@/entities/rating";
import { teamApi } from "@/entities/team";
import { useAuth } from "@/entities/session";
import { profileApi } from "@/entities/profile";
import { referenceDataApi } from "@/entities/reference-data";
import { municipalityApi } from "@/entities/municipality";
import { SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Checkbox } from "@/shared/ui/checkbox";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { useNotifications } from "@/shared/lib/notifications";

export function RatingView() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const router = useRouter();
  const [year, setYear] = useState(new Date().getFullYear());
  const [disciplineId, setDisciplineId] = useState(0);
  const [municipalityId, setMunicipalityId] = useState(0);
  const [offset, setOffset] = useState(0);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const load = useCallback(() => ratingApi.list({ year, discipline_id: disciplineId || undefined, municipality_id: municipalityId || undefined, limit: 50, offset }), [year, disciplineId, municipalityId, offset]);
  const state = useAsyncData(load);
  const references = useAsyncData(useCallback(async () => {
    const [disciplines, municipalities] = await Promise.all([referenceDataApi.disciplines(), municipalityApi.list({ limit: 100 })]);
    return { disciplines, municipalities };
  }, []));
  const profile = useAsyncData(useCallback(() => profileApi.me(), []));
  const canFormTeam = Boolean(identity?.permissions?.includes("teams.manage") && profile.data?.roles.some((role) => role.code === "admin" || role.code === "super_admin"));

  function toggleMember(id: number, selected: boolean) {
    setSelectedIds((current) => selected ? [...new Set([...current, id])] : current.filter((item) => item !== id));
  }

  async function createTeam(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedIds.length) return;
    const name = String(new FormData(event.currentTarget).get("name") ?? "").trim();
    setPending(true);
    try {
      await teamApi.create({ name, member_ids: selectedIds, coach_ids: [] });
      notify.success("Сборная создана из выбранных спортсменов");
      setCreateOpen(false);
      setSelectedIds([]);
      router.push("/workspace/teams");
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось создать сборную"); }
    finally { setPending(false); }
  }

  return <div className="mx-auto w-full max-w-6xl">
    <SectionHeading title="Рейтинг спортсменов" description="Баллы за завершённые соревнования выбранного года." />
    <div className="mb-5 grid gap-3 sm:grid-cols-3">
      <Input aria-label="Год рейтинга" type="number" min={2000} max={new Date().getFullYear() + 1} value={year} onChange={(event) => { setYear(Number(event.target.value)); setOffset(0); setSelectedIds([]); }} />
      <Select value={String(disciplineId)} onValueChange={(value) => { setDisciplineId(Number(value)); setOffset(0); setSelectedIds([]); }}><SelectTrigger><SelectValue placeholder="Все дисциплины" /></SelectTrigger><SelectContent><SelectItem value="0">Все дисциплины</SelectItem>{references.data?.disciplines.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
      <Select value={String(municipalityId)} onValueChange={(value) => { setMunicipalityId(Number(value)); setOffset(0); setSelectedIds([]); }}><SelectTrigger><SelectValue placeholder="Все муниципалитеты" /></SelectTrigger><SelectContent><SelectItem value="0">Все муниципалитеты</SelectItem>{references.data?.municipalities.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
    </div>
    {state.loading ? <DataLoading rows={5} /> : null}
    {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
    {!state.loading && !state.error && !state.data?.items.length ? <DataEmpty title="Результатов нет" description="Попробуйте другой год или фильтр." /> : null}
    <div className="grid gap-2">{state.data?.items.map((entry) => <Card key={entry.user_id} className="flex flex-wrap items-center justify-between gap-3 bg-white/75 p-4 shadow-none"><div className="flex min-w-0 items-center gap-3">{canFormTeam ? <Checkbox checked={selectedIds.includes(entry.user_id)} onCheckedChange={(checked) => toggleMember(entry.user_id, checked === true)} aria-label={`Выбрать ${entry.last_name} ${entry.first_name} в сборную`} /> : null}<div className="min-w-0"><p className="font-medium">{entry.position}. <Link className="hover:underline focus-visible:underline" href={`/profile/${entry.user_id}`}>{entry.last_name} {entry.first_name} {entry.middle_name ?? ""}</Link></p><p className="text-sm text-muted-foreground">{entry.municipality_name ?? "Муниципалитет не указан"} · {entry.result_count} результатов · {entry.wins} побед</p></div></div><p className="font-medium tabular-nums">{entry.total_points} баллов</p></Card>)}</div>
    {canFormTeam ? <div className="mt-4 flex flex-wrap items-center gap-3"><Button disabled={!selectedIds.length} onClick={() => setCreateOpen(true)}>Создать сборную из выбранных</Button><p className="text-sm text-muted-foreground">Выбрано: {selectedIds.length}. Выбор сохраняется при переходе между страницами рейтинга.</p></div> : null}
    <div className="mt-4 flex gap-2"><Button variant="outline" disabled={offset === 0 || state.loading} onClick={() => setOffset(Math.max(0, offset - 50))}>Назад</Button><Button variant="outline" disabled={state.loading || (state.data?.items.length ?? 0) < 50} onClick={() => setOffset(offset + 50)}>Далее</Button></div>
    <Dialog open={createOpen} onOpenChange={setCreateOpen}><DialogContent><DialogHeader><DialogTitle>Новая сборная</DialogTitle><DialogDescription>В состав войдут {selectedIds.length} выбранных спортсменов. После создания откроется страница сборных.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createTeam}><label htmlFor="rating-team-name" className="text-sm font-medium">Название сборной</label><Input id="rating-team-name" name="name" required maxLength={200} defaultValue={`Сборная ${year}`} /><DialogFooter><Button type="submit" disabled={pending}>{pending ? "Создание…" : "Создать сборную"}</Button></DialogFooter></form></DialogContent></Dialog>
  </div>;
}
