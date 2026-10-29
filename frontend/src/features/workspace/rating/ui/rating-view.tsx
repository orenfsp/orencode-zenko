"use client";

import { useCallback, useMemo, useState, type FormEvent, type ReactNode } from "react";
import Link from "next/link";
import { Download, Info, LoaderCircle, LockKeyhole, Save, UsersRound } from "lucide-react";

import { ratingApi } from "@/entities/rating";
import { municipalityApi } from "@/entities/municipality";
import { referenceDataApi } from "@/entities/reference-data";
import { registryApi } from "@/entities/registry";
import { useAuth } from "@/entities/session";
import { useIsMobile } from "@/hooks/use-mobile";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { downloadBlob } from "@/shared/lib/download";
import { useNotifications } from "@/shared/lib/notifications";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { ChartBars } from "@/shared/ui/chart";
import { Checkbox } from "@/shared/ui/checkbox";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Drawer, DrawerContent, DrawerDescription, DrawerFooter, DrawerHandle, DrawerHeader, DrawerTitle } from "@/shared/ui/drawer";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

function ResponsiveSurface({ open, onOpenChange, title, description, children }: { open: boolean; onOpenChange: (open: boolean) => void; title: string; description: string; children: ReactNode }) {
  const isMobile = useIsMobile();
  if (isMobile) {
    return <Drawer open={open} onOpenChange={onOpenChange}><DrawerContent><DrawerHandle /><DrawerHeader><DrawerTitle>{title}</DrawerTitle><DrawerDescription>{description}</DrawerDescription></DrawerHeader>{children}</DrawerContent></Drawer>;
  }
  return <Dialog open={open} onOpenChange={onOpenChange}><DialogContent><DialogHeader><DialogTitle>{title}</DialogTitle><DialogDescription>{description}</DialogDescription></DialogHeader>{children}</DialogContent></Dialog>;
}

export function RatingView() {
  const currentYear = new Date().getFullYear();
  const { identity } = useAuth();
  const notify = useNotifications();
  const permissions = useMemo(() => new Set(identity?.permissions ?? []), [identity?.permissions]);
  const canManageSelection = permissions.has("rating.manage");
  const canReadRegistry = permissions.has("registry.read");
  const [year, setYear] = useState(currentYear);
  const [disciplineId, setDisciplineId] = useState(0);
  const [ageGroupId, setAgeGroupId] = useState(0);
  const [municipalityId, setMunicipalityId] = useState(0);
  const [offset, setOffset] = useState(0);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [captainId, setCaptainId] = useState(0);
  const [coachId, setCoachId] = useState(0);
  const [draftContext, setDraftContext] = useState("");
  const [search, setSearch] = useState("");
  const [selectionOpen, setSelectionOpen] = useState(false);
  const [lockOpen, setLockOpen] = useState(false);
  const [pending, setPending] = useState<"save" | "lock" | "order" | null>(null);

  const state = useAsyncData(useCallback(() => ratingApi.list({ year, discipline_id: disciplineId || undefined, age_group_id: ageGroupId || undefined, municipality_id: municipalityId || undefined, limit: 50, offset }), [year, disciplineId, ageGroupId, municipalityId, offset]));
  const references = useAsyncData(useCallback(async () => {
    const [disciplines, ageGroups, municipalities] = await Promise.all([referenceDataApi.disciplines(), referenceDataApi.ageGroups(), municipalityApi.list({ limit: 100 })]);
    return { disciplines, ageGroups, municipalities };
  }, []));
  const coaches = useAsyncData(useCallback(async () => canReadRegistry ? (await registryApi.users({ role: "coach", limit: 100 })).items : [], [canReadRegistry]));
  const selectionState = useAsyncData(useCallback(async () => {
    if (!disciplineId || !ageGroupId) return null;
    return ratingApi.selection({ year, discipline_id: disciplineId, age_group_id: ageGroupId });
  }, [year, disciplineId, ageGroupId]));
  const selection = selectionState.data;
  const canEditSelection = canManageSelection && selection?.status !== "locked";
  const currentContext = `${year}:${disciplineId}:${ageGroupId}`;
  const draftMatchesContext = draftContext === currentContext;
  const selectionMemberIds = useMemo(() => selection?.members.map((member) => member.user_id) ?? [], [selection?.members]);
  const effectiveSelectedIds = useMemo(() => draftMatchesContext ? selectedIds : selectionMemberIds, [draftMatchesContext, selectedIds, selectionMemberIds]);
  const effectiveCaptainId = draftMatchesContext ? captainId : selection?.captain_user_id ?? 0;
  const effectiveCoachId = draftMatchesContext ? coachId : selection?.coach_user_id ?? 0;

  const visibleItems = useMemo(() => {
    const query = search.trim().toLocaleLowerCase("ru");
    return (state.data?.items ?? []).filter((entry) => `${entry.last_name} ${entry.first_name} ${entry.middle_name ?? ""} ${entry.municipality_name ?? ""}`.toLocaleLowerCase("ru").includes(query));
  }, [search, state.data?.items]);
  const captainOptions = useMemo(() => effectiveSelectedIds.map((userId) => {
    const ratingEntry = (state.data?.items ?? []).find((item) => item.user_id === userId);
    if (ratingEntry) return { id: userId, name: `${ratingEntry.last_name} ${ratingEntry.first_name}` };
    const selectionMember = selection?.members.find((item) => item.user_id === userId);
    return selectionMember ? { id: userId, name: `${selectionMember.last_name} ${selectionMember.first_name}` } : null;
  }).filter((item): item is { id: number; name: string } => item !== null), [effectiveSelectedIds, selection?.members, state.data?.items]);
  const hasFilters = year !== currentYear || disciplineId !== 0 || ageGroupId !== 0 || municipalityId !== 0;
  const selectionContextReady = disciplineId !== 0 && ageGroupId !== 0;

  function clearSelection() { setSelectedIds([]); setCaptainId(0); setCoachId(0); setDraftContext(""); }
  function resetFilters() { setYear(currentYear); setDisciplineId(0); setAgeGroupId(0); setMunicipalityId(0); setOffset(0); clearSelection(); }
  function toggleMember(id: number, selected: boolean) {
    setDraftContext(currentContext);
    setSelectedIds(selected ? [...new Set([...effectiveSelectedIds, id])] : effectiveSelectedIds.filter((item) => item !== id));
    setCaptainId(!selected && effectiveCaptainId === id ? 0 : effectiveCaptainId);
    setCoachId(effectiveCoachId);
  }
  function openSelectionEditor() {
    setDraftContext(currentContext);
    setSelectedIds(effectiveSelectedIds);
    setCaptainId(effectiveCaptainId);
    setCoachId(effectiveCoachId);
    setSelectionOpen(true);
  }

  async function saveSelection(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectionContextReady || !effectiveSelectedIds.length) return;
    setPending("save");
    try {
      await ratingApi.saveSelection({ season_year: year, discipline_id: disciplineId, age_group_id: ageGroupId, member_ids: effectiveSelectedIds, captain_user_id: effectiveCaptainId || null, coach_user_id: effectiveCoachId || null });
      notify.success("Черновик сезонного состава сохранён");
      setSelectionOpen(false);
      selectionState.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось сохранить состав"); }
    finally { setPending(null); }
  }

  async function lockSelection() {
    if (!selection) return;
    setPending("lock");
    try {
      await ratingApi.lockSelection(selection.id);
      notify.success("Сезонный состав зафиксирован");
      setLockOpen(false);
      selectionState.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось зафиксировать состав"); }
    finally { setPending(null); }
  }

  async function downloadOrder() {
    if (!selectionContextReady) return;
    setPending("order");
    try {
      const order = await ratingApi.order({ year, discipline_id: disciplineId, age_group_id: ageGroupId });
      downloadBlob(order.blob, order.filename);
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось сформировать проект приказа"); }
    finally { setPending(null); }
  }

  const selectionForm = <form className="grid min-h-0 gap-4 overflow-y-auto px-4 pb-4 md:overflow-visible md:px-0 md:pb-0" onSubmit={saveSelection}>
    <div className="grid gap-2"><label className="text-sm font-medium" htmlFor="rating-captain">Капитан</label><Select value={String(effectiveCaptainId)} onValueChange={(value) => { setDraftContext(currentContext); setCaptainId(Number(value)); }}><SelectTrigger id="rating-captain"><SelectValue placeholder="Не назначен" /></SelectTrigger><SelectContent><SelectItem value="0">Не назначен</SelectItem>{captainOptions.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select></div>
    <div className="grid gap-2"><label className="text-sm font-medium" htmlFor="rating-coach">Тренер состава</label><Select value={String(effectiveCoachId)} onValueChange={(value) => { setDraftContext(currentContext); setCoachId(Number(value)); }} disabled={!canReadRegistry}><SelectTrigger id="rating-coach"><SelectValue placeholder="Не назначен" /></SelectTrigger><SelectContent><SelectItem value="0">Не назначен</SelectItem>{(coaches.data ?? []).map((coach) => <SelectItem key={coach.id} value={String(coach.id)}>{coach.last_name} {coach.first_name}</SelectItem>)}</SelectContent></Select>{!canReadRegistry ? <p className="text-xs text-muted-foreground">Назначить тренера может пользователь с доступом к реестру.</p> : null}</div>
    <p className="text-sm text-muted-foreground">Порядок спортсменов сохраняется в порядке выбора. После сохранения состав останется черновиком.</p>
    <DialogFooter className="hidden md:flex"><Button type="submit" disabled={pending !== null}><Save />{pending === "save" ? "Сохранение…" : "Сохранить черновик"}</Button></DialogFooter>
    <DrawerFooter className="p-0 md:hidden"><Button type="submit" disabled={pending !== null}><Save />{pending === "save" ? "Сохранение…" : "Сохранить черновик"}</Button></DrawerFooter>
  </form>;

  return <div className="mx-auto w-full max-w-6xl">
    <WorkspaceToolbar search={search} onSearchChange={setSearch} searchPlaceholder="Найти спортсмена" filterActive={hasFilters} onResetFilters={resetFilters} filterContent={<div className="grid gap-3">
      <Input aria-label="Год рейтинга" type="number" min={2000} max={currentYear + 1} value={year} onChange={(event) => { setYear(Number(event.target.value)); setOffset(0); clearSelection(); }} />
      <Select value={String(disciplineId)} onValueChange={(value) => { setDisciplineId(Number(value)); setOffset(0); clearSelection(); }}><SelectTrigger><SelectValue placeholder="Все дисциплины" /></SelectTrigger><SelectContent><SelectItem value="0">Все дисциплины</SelectItem>{references.data?.disciplines.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
      <Select value={String(ageGroupId)} onValueChange={(value) => { setAgeGroupId(Number(value)); setOffset(0); clearSelection(); }}><SelectTrigger><SelectValue placeholder="Все возрастные группы" /></SelectTrigger><SelectContent><SelectItem value="0">Все возрастные группы</SelectItem>{references.data?.ageGroups.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
      <Select value={String(municipalityId)} onValueChange={(value) => { setMunicipalityId(Number(value)); setOffset(0); clearSelection(); }}><SelectTrigger><SelectValue placeholder="Все муниципалитеты" /></SelectTrigger><SelectContent><SelectItem value="0">Все муниципалитеты</SelectItem>{references.data?.municipalities.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
    </div>} onCreate={canEditSelection && selectionContextReady && effectiveSelectedIds.length ? openSelectionEditor : undefined} createLabel={selection ? "Изменить состав" : "Сохранить состав"} createClassName="hidden sm:inline-flex" />

    {canManageSelection && !selectionContextReady ? <Alert className="mb-4 bg-card/70"><Info /><AlertTitle>Чтобы сформировать состав, уточните рейтинг</AlertTitle><AlertDescription>Выберите дисциплину и возрастную группу в фильтрах. После этого можно отметить спортсменов.</AlertDescription></Alert> : null}
    {selectionContextReady && !selectionState.loading ? <Card variant="glass" className="mb-4 p-4 sm:p-5"><div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><div><p className="font-medium">Сезонный состав · {selection?.status === "locked" ? "зафиксирован" : selection ? "черновик" : "не сформирован"}</p><p className="mt-1 text-sm text-muted-foreground">{selection ? `${selection.members.length} спортсменов · ${selection.discipline.name} · ${selection.age_group.name}` : "Отметьте спортсменов в рейтинге и сохраните черновик состава."}</p></div>{selection ? <div className="flex flex-wrap gap-2"><Button variant="outline" disabled={pending !== null} onClick={() => void downloadOrder()}>{pending === "order" ? <LoaderCircle className="animate-spin" /> : <Download />}Проект приказа</Button>{canManageSelection && selection.status === "draft" ? <Button variant="outline" disabled={pending !== null} onClick={() => setLockOpen(true)}><LockKeyhole />Зафиксировать состав</Button> : null}</div> : null}</div></Card> : null}
    {state.loading ? <DataLoading rows={5} /> : null}
    {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
    {!state.loading && !state.error && !visibleItems.length ? <DataEmpty title={search || hasFilters ? "Спортсмены не найдены" : "Результатов рейтинга пока нет"} description={search || hasFilters ? "Измените запрос или сбросьте фильтры." : "Рейтинг появится после публикации результатов соревнований."} action={search || hasFilters ? <Button variant="outline" size="sm" onClick={() => { setSearch(""); resetFilters(); }}>Сбросить фильтры</Button> : undefined} /> : null}
    {!state.loading && !state.error && (state.data?.items.length ?? 0) > 1 ? <Card variant="glass" className="mb-4 p-5"><CardHeader className="p-0"><CardTitle>Рейтинг за {year} год</CardTitle><CardDescription>Баллы начисляются по результатам: используется итоговый score, а при его отсутствии — 100, 80, 60, 50, 40, 30, 20 или 10 баллов за место.</CardDescription></CardHeader><div className="mt-4"><ChartBars points={(state.data?.items ?? []).slice(0, 8).map((entry) => ({ label: `${entry.last_name} ${entry.first_name.slice(0, 1)}.`, value: entry.total_points }))} label="Столбчатая диаграмма баллов спортсменов в рейтинге" /></div></Card> : null}
    <div className="grid gap-3">{visibleItems.map((entry) => <Card key={entry.user_id} variant="glass" className="p-4 sm:p-5"><div className="flex items-start gap-3"><div className="flex min-w-0 flex-1 items-start gap-3">{canEditSelection && selectionContextReady ? <Checkbox className="mt-1" checked={effectiveSelectedIds.includes(entry.user_id)} onCheckedChange={(checked) => toggleMember(entry.user_id, checked === true)} aria-label={`Добавить ${entry.last_name} ${entry.first_name} в сезонный состав`} /> : null}<div className="min-w-0"><p className="text-xs text-muted-foreground">{entry.position} место</p><Link className="font-medium hover:underline focus-visible:underline" href={`/profile/${entry.user_id}`}>{entry.last_name} {entry.first_name} {entry.middle_name ?? ""}</Link><p className="mt-1 text-sm text-muted-foreground">{entry.municipality_name ?? "Муниципалитет не указан"}</p></div></div><p className="shrink-0 text-right text-lg font-semibold tabular-nums">{entry.total_points}<span className="block text-xs font-normal text-muted-foreground">баллов</span></p></div><div className="mt-3 flex gap-4 pl-7 text-xs text-muted-foreground"><span>{entry.result_count} результатов</span><span>{entry.wins} побед</span></div></Card>)}</div>
    {canEditSelection && selectionContextReady && effectiveSelectedIds.length ? <p className="mt-3 text-sm text-muted-foreground">В сезонный состав выбрано: {effectiveSelectedIds.length}</p> : null}
    {canEditSelection && selectionContextReady && effectiveSelectedIds.length ? <Button className="fixed inset-x-4 bottom-4 z-30 h-11 shadow-lg sm:hidden" onClick={openSelectionEditor}><UsersRound />{selection ? "Изменить состав" : "Сохранить состав"}</Button> : null}
    <div className="mt-4 flex gap-2"><Button variant="outline" disabled={offset === 0 || state.loading} onClick={() => setOffset(Math.max(0, offset - 50))}>Назад</Button><Button variant="outline" disabled={state.loading || (state.data?.items.length ?? 0) < 50} onClick={() => setOffset(offset + 50)}>Далее</Button></div>
    <ResponsiveSurface open={selectionOpen} onOpenChange={setSelectionOpen} title="Сезонный состав" description={`Выбрано спортсменов: ${effectiveSelectedIds.length}. Капитана и тренера можно назначить сейчас или позже.`}>{selectionForm}</ResponsiveSurface>
    <ResponsiveSurface open={lockOpen} onOpenChange={setLockOpen} title="Зафиксировать состав?" description="После фиксации изменить порядок и участников через этот интерфейс будет нельзя."><div className="grid gap-4 px-4 pb-4 md:px-0 md:pb-0"><Alert><LockKeyhole /><AlertTitle>Это действие нельзя повторить</AlertTitle><AlertDescription>Проверьте участников, капитана и тренера до фиксации.</AlertDescription></Alert><DialogFooter className="hidden md:flex"><Button variant="outline" onClick={() => setLockOpen(false)}>Вернуться к проверке</Button><Button onClick={() => void lockSelection()} disabled={pending !== null}>{pending === "lock" ? "Фиксация…" : "Зафиксировать"}</Button></DialogFooter><DrawerFooter className="p-0 md:hidden"><Button variant="outline" onClick={() => setLockOpen(false)}>Вернуться к проверке</Button><Button onClick={() => void lockSelection()} disabled={pending !== null}>{pending === "lock" ? "Фиксация…" : "Зафиксировать"}</Button></DrawerFooter></div></ResponsiveSurface>
  </div>;
}
