"use client";

import { useCallback, useMemo, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ratingApi } from "@/entities/rating";
import { teamApi } from "@/entities/team";
import { useAuth } from "@/entities/session";
import { profileApi } from "@/entities/profile";
import { referenceDataApi } from "@/entities/reference-data";
import { municipalityApi } from "@/entities/municipality";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Checkbox } from "@/shared/ui/checkbox";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { useNotifications } from "@/shared/lib/notifications";
import { ChartBars } from "@/shared/ui/chart";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

export function RatingView() {
  const currentYear = new Date().getFullYear();
  const { identity } = useAuth();
  const notify = useNotifications();
  const router = useRouter();
  const [year, setYear] = useState(currentYear);
  const [disciplineId, setDisciplineId] = useState(0);
  const [municipalityId, setMunicipalityId] = useState(0);
  const [offset, setOffset] = useState(0);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [search, setSearch] = useState("");
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
  const visibleItems = useMemo(() => {
    const query = search.trim().toLocaleLowerCase("ru");
    return (state.data?.items ?? []).filter((entry) => `${entry.last_name} ${entry.first_name} ${entry.middle_name ?? ""} ${entry.municipality_name ?? ""}`.toLocaleLowerCase("ru").includes(query));
  }, [search, state.data?.items]);

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
    <WorkspaceToolbar
      search={search}
      onSearchChange={setSearch}
      searchPlaceholder="Найти спортсмена"
      filterActive={year !== currentYear || disciplineId !== 0 || municipalityId !== 0}
      onResetFilters={() => { setYear(currentYear); setDisciplineId(0); setMunicipalityId(0); setOffset(0); setSelectedIds([]); }}
      filterContent={<div className="grid gap-3"><Input aria-label="Год рейтинга" type="number" min={2000} max={new Date().getFullYear() + 1} value={year} onChange={(event) => { setYear(Number(event.target.value)); setOffset(0); setSelectedIds([]); }} /><Select value={String(disciplineId)} onValueChange={(value) => { setDisciplineId(Number(value)); setOffset(0); setSelectedIds([]); }}><SelectTrigger><SelectValue placeholder="Все дисциплины" /></SelectTrigger><SelectContent><SelectItem value="0">Все дисциплины</SelectItem>{references.data?.disciplines.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select><Select value={String(municipalityId)} onValueChange={(value) => { setMunicipalityId(Number(value)); setOffset(0); setSelectedIds([]); }}><SelectTrigger><SelectValue placeholder="Все муниципалитеты" /></SelectTrigger><SelectContent><SelectItem value="0">Все муниципалитеты</SelectItem>{references.data?.municipalities.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select></div>}
      onCreate={canFormTeam && selectedIds.length ? () => setCreateOpen(true) : undefined}
      createLabel="Создать сборную"
      createClassName="hidden sm:inline-flex"
    />
    {state.loading ? <DataLoading rows={5} /> : null}
    {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
    {!state.loading && !state.error && !visibleItems.length ? <DataEmpty title={search || year !== currentYear || disciplineId !== 0 || municipalityId !== 0 ? "Спортсмены не найдены" : "Результатов нет"} description={search || year !== currentYear || disciplineId !== 0 || municipalityId !== 0 ? "Измените запрос или сбросьте фильтры." : "Результаты появятся после публикации."} action={search || year !== currentYear || disciplineId !== 0 || municipalityId !== 0 ? <Button variant="outline" size="sm" onClick={() => { setSearch(""); setYear(currentYear); setDisciplineId(0); setMunicipalityId(0); setOffset(0); }}>Сбросить фильтры</Button> : undefined} /> : null}
    {!state.loading && !state.error && (state.data?.items.length ?? 0) > 1 ? <Card variant="glass" className="mb-4 p-5"><CardHeader className="p-0"><CardTitle>Текущий рейтинг</CardTitle><CardDescription>Топ спортсменов на выбранной странице рейтинга за {year} год.</CardDescription></CardHeader><div className="mt-4"><ChartBars points={(state.data?.items ?? []).slice(0, 8).map((entry) => ({ label: `${entry.last_name} ${entry.first_name.slice(0, 1)}.`, value: entry.total_points }))} label="Столбчатая диаграмма баллов спортсменов в рейтинге" /></div></Card> : null}
    <div className="grid gap-3">
      {visibleItems.map((entry) => <Card key={entry.user_id} variant="glass" className="p-4 sm:p-5"><div className="flex items-start gap-3"><div className="flex min-w-0 flex-1 items-start gap-3">{canFormTeam ? <Checkbox className="mt-1" checked={selectedIds.includes(entry.user_id)} onCheckedChange={(checked) => toggleMember(entry.user_id, checked === true)} aria-label={`Выбрать ${entry.last_name} ${entry.first_name} в сборную`} /> : null}<div className="min-w-0"><p className="text-xs text-muted-foreground">{entry.position} место</p><Link className="font-medium hover:underline focus-visible:underline" href={`/profile/${entry.user_id}`}>{entry.last_name} {entry.first_name} {entry.middle_name ?? ""}</Link><p className="mt-1 text-sm text-muted-foreground">{entry.municipality_name ?? "Муниципалитет не указан"}</p></div></div><p className="shrink-0 text-right text-lg font-semibold tabular-nums">{entry.total_points}<span className="block text-xs font-normal text-muted-foreground">баллов</span></p></div><div className="mt-3 flex gap-4 pl-7 text-xs text-muted-foreground"><span>{entry.result_count} результатов</span><span>{entry.wins} побед</span></div></Card>)}
    </div>
    {canFormTeam && selectedIds.length ? <p className="mt-3 text-sm text-muted-foreground">Выбрано спортсменов: {selectedIds.length}. Используйте кнопку создания сборной.</p> : null}
    {canFormTeam && selectedIds.length ? <Button className="fixed inset-x-4 bottom-4 z-30 h-11 shadow-lg sm:hidden" onClick={() => setCreateOpen(true)}>Создать сборную из выбранных</Button> : null}
    <div className="mt-4 flex gap-2"><Button variant="outline" disabled={offset === 0 || state.loading} onClick={() => setOffset(Math.max(0, offset - 50))}>Назад</Button><Button variant="outline" disabled={state.loading || (state.data?.items.length ?? 0) < 50} onClick={() => setOffset(offset + 50)}>Далее</Button></div>
    <Dialog open={createOpen} onOpenChange={setCreateOpen}><DialogContent><DialogHeader><DialogTitle>Новая сборная</DialogTitle><DialogDescription>В состав войдут {selectedIds.length} выбранных спортсменов. После создания откроется страница сборных.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createTeam}><label htmlFor="rating-team-name" className="text-sm font-medium">Название сборной</label><Input id="rating-team-name" name="name" required maxLength={200} defaultValue={`Сборная ${year}`} /><DialogFooter><Button type="submit" disabled={pending}>{pending ? "Создание…" : "Создать сборную"}</Button></DialogFooter></form></DialogContent></Dialog>
  </div>;
}
