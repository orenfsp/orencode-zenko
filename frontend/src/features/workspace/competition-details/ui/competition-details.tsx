"use client";

import { FormEvent, useCallback, useState } from "react";
import {
  ArrowLeft,
  CalendarDays,
  ExternalLink,
  FileText,
  LoaderCircle,
  MapPin,
  Pencil,
  Plus,
  Trash2,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { competitionApi } from "@/entities/competition";
import { competitionApplicationApi } from "@/entities/competition-application";
import { competitionEventApi, type CompetitionEvent } from "@/entities/competition-event";
import { competitionJudgeApi } from "@/entities/competition-judge";
import { competitionResultApi, type CompetitionResultTable } from "@/entities/competition-result";
import { profileApi } from "@/entities/profile";
import { referenceDataApi } from "@/entities/reference-data";
import { registryApi } from "@/entities/registry";
import { userCoachApi } from "@/entities/user-coach";
import { useAuth } from "@/entities/session";
import { formatDate, fullName, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { DataError, DataLoading } from "@/shared/ui/data-state";
import { useNotifications } from "@/shared/lib/notifications";
import { downloadBlob } from "@/shared/lib/download";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Textarea } from "@/shared/ui/textarea";
import { Checkbox } from "@/shared/ui/checkbox";
import { ConfirmAction } from "@/shared/ui/confirm-action";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { ScrollMask } from "@/shared/ui/scroll-mask";
import { ResultEntryEditor } from "./result-entry-editor";

interface EventEditor {
  mode: "create" | "edit";
  eventId?: number;
  ageGroupId: number | null;
}

export function CompetitionDetails({ competitionId }: { competitionId: number }) {
  const { identity } = useAuth();
  const notify = useNotifications();
  const router = useRouter();
  const permissions = new Set(identity?.permissions ?? []);
  const canUpdate = permissions.has("competitions.update");
  const canManageJudges = permissions.has("competition_judges.manage");
  const canReadRegistry = permissions.has("registry.read");
  const canReadResults = permissions.has("competition_results.read");
  const canManageResults = permissions.has("competition_results.manage");
  const canApproveResults = permissions.has("competition_results.approve");
  const canReadEvents = permissions.has("competition_events.read");
  const canCreateEvents = permissions.has("competition_events.create");
  const canUpdateEvents = permissions.has("competition_events.update");
  const canDeleteEvents = permissions.has("competition_events.delete");
  const [pending, setPending] = useState(false);
  const [result, setResult] = useState<CompetitionResultTable | null>(null);
  const [resultEventId, setResultEventId] = useState<number | null>(null);
  const [resultJson, setResultJson] = useState("");
  const [resultInputError, setResultInputError] = useState<string | null>(null);
  const [resultVersion, setResultVersion] = useState(0);
  const [applyingEventId, setApplyingEventId] = useState<number | null>(null);
  const [eventEditor, setEventEditor] = useState<EventEditor | null>(null);
  const [deletingEvent, setDeletingEvent] = useState<CompetitionEvent | null>(null);
  const load = useCallback(
    async () => {
      const managedCompetition = canUpdate ? await competitionApi.adminGet(competitionId).catch(() => null) : null;
      const [competition, events, judges, profile, ageGroups, judgePositions, judgeUsersPage] = await Promise.all([
        managedCompetition ? Promise.resolve(managedCompetition) : competitionApi.get(competitionId),
        canReadEvents ? competitionEventApi.adminList(competitionId) : competitionEventApi.publicList(competitionId),
        canManageJudges ? competitionJudgeApi.adminList(competitionId) : competitionJudgeApi.publicList(competitionId),
        profileApi.me(),
        canCreateEvents || canUpdateEvents ? referenceDataApi.ageGroups() : Promise.resolve([]),
        canManageJudges ? competitionJudgeApi.positions() : Promise.resolve([]),
        canManageJudges && canReadRegistry ? registryApi.adminUsers({ role: "judge", limit: 100 }) : Promise.resolve(null),
      ]);
      const activeRole = profile.roles.find((role) => role.id === identity?.roleId)?.code;
      const athletes = activeRole === "coach" ? await userCoachApi.byCoach({ limit: 100 }) : [];
      const athleteIds = [...new Set(athletes.map((item) => item.participant_user_id))];
      const athleteProfiles = await Promise.allSettled(athleteIds.map((userId) => profileApi.public(userId)));
      const athleteNames = Object.fromEntries(athleteProfiles.flatMap((item, index) => item.status === "fulfilled" ? [[athleteIds[index], fullName(item.value.last_name, item.value.first_name, item.value.middle_name)]] : []));
      return { competition, events, judges, activeRole, athletes, ageGroups, judgePositions, judgeUsers: judgeUsersPage?.items ?? [], canManageCompetition: managedCompetition !== null, athleteNames };
    },
    [canCreateEvents, canManageJudges, canReadEvents, canReadRegistry, canUpdate, canUpdateEvents, competitionId, identity?.roleId],
  );
  const state = useAsyncData(load);

  async function run(action: () => Promise<unknown>, success: string) {
    setPending(true);
    try { await action(); notify.success(success); state.reload(); }
    catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось выполнить действие"); }
    finally { setPending(false); }
  }

  function updateCompetition(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(() => competitionApi.update(competitionId, { name: String(form.get("name") ?? ""), description: String(form.get("description") ?? "") || null }), "Соревнование обновлено");
  }

  function transitionCompetition(action: "submit" | "publish" | "cancel" | "return") {
    const actions = {
      submit: () => competitionApi.submit(competitionId),
      publish: () => competitionApi.publish(competitionId),
      cancel: () => competitionApi.cancel(competitionId),
      return: () => competitionApi.returnToDraft(competitionId),
    };
    const messages = { submit: "Соревнование отправлено на проверку", publish: "Соревнование опубликовано", cancel: "Соревнование отменено", return: "Соревнование возвращено на доработку" };
    void run(actions[action], messages[action]);
  }

  function assignJudge(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(() => competitionJudgeApi.assign(competitionId, { judge_user_id: Number(form.get("judge_user_id")), position_id: Number(form.get("position_id")) }), "Судья назначен");
  }

  async function applyForEvent(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!applyingEventId || !state.data) return;
    const athleteId = Number(new FormData(event.currentTarget).get("athlete_user_id"));
    setPending(true);
    try {
      await competitionApplicationApi.create({ event_id: applyingEventId, athlete_user_id: state.data.activeRole === "coach" ? athleteId : undefined });
      notify.success("Заявка подана. Статус доступен в разделе «Мои заявки»");
      setApplyingEventId(null);
      router.push("/workspace/applications");
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось подать заявку"); }
    finally { setPending(false); }
  }

  function saveEvent(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!eventEditor) return;
    const ageGroupId = eventEditor.ageGroupId;
    const action = eventEditor.mode === "create"
      ? () => competitionEventApi.create(competitionId, { age_group_id: ageGroupId })
      : eventEditor.eventId
        ? () => competitionEventApi.update(competitionId, eventEditor.eventId!, { age_group_id: ageGroupId })
        : null;
    if (!action) return;
    void run(action, eventEditor.mode === "create" ? "Зачёт создан" : "Зачёт обновлён");
    setEventEditor(null);
  }

  function removeEvent() {
    if (!deletingEvent) return;
    void run(() => competitionEventApi.remove(competitionId, deletingEvent.id), "Зачёт удалён");
    setDeletingEvent(null);
  }

  async function openResults(eventId: number) {
    setPending(true);
    try {
      if (permissions.has("competition_events.read")) await competitionEventApi.adminGet(competitionId, eventId);
      else await competitionEventApi.publicGet(competitionId, eventId);
      const table = await competitionResultApi.get(competitionId, eventId);
      setResult(table);
      setResultVersion((version) => version + 1);
      setResultEventId(eventId);
      setResultJson(JSON.stringify({ entries: table.entries.map((entry) => ({ entry_id: entry.id, result_status: entry.result_status, tasks: entry.tasks.map((task) => ({ task_number: task.task_number, status: task.status, solve_time_deciseconds: task.solve_time_deciseconds, negative_attempts: task.negative_attempts })) })) }, null, 2));
      setResultInputError(null);
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить результаты"); }
    finally { setPending(false); }
  }

  async function resultAction(action: () => Promise<CompetitionResultTable | unknown>, success: string) {
    setPending(true);
    try {
      const next = await action();
      if (next && typeof next === "object" && "entries" in next) {
        const table = next as CompetitionResultTable;
        setResult(table);
        setResultVersion((version) => version + 1);
        setResultJson(JSON.stringify({ entries: table.entries.map((entry) => ({ entry_id: entry.id, result_status: entry.result_status, tasks: entry.tasks.map((task) => ({ task_number: task.task_number, status: task.status, solve_time_deciseconds: task.solve_time_deciseconds, negative_attempts: task.negative_attempts })) })) }, null, 2));
      } else if (next && typeof next === "object" && "status" in next && next.status === "approved") {
        setResult((current) => current ? { ...current, status: "approved", approved_at: "approved_at" in next ? String(next.approved_at) : current.approved_at } : current);
      }
      notify.success(success);
      state.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось обновить результаты"); }
    finally { setPending(false); }
  }

  function configureResults(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!resultEventId || !result) return;
    const form = new FormData(event.currentTarget);
    const mode = String(form.get("mode"));
    let teams: unknown;
    try { teams = JSON.parse(String(form.get("teams") || "[]")); }
    catch { setResultInputError("Проверьте синтаксис списка команд."); return; }
    if (!Array.isArray(teams)) { setResultInputError("Составы команд должны быть массивом."); return; }
    const confirmedReset = form.get("confirm_reset") === "on";
    if (result.entries.length && !confirmedReset) {
      setResultInputError("Повторная настройка удалит введённые результаты. Отметьте подтверждение сброса.");
      return;
    }
    setResultInputError(null);
    void resultAction(() => competitionResultApi.configure(competitionId, resultEventId, { mode, task_count: Number(form.get("task_count")), teams: mode === "team" ? teams : undefined, confirm_reset: confirmedReset }), "Настройки сохранены");
  }

  function saveResultDraft() {
    if (!resultEventId) return;
    let parsed: unknown;
    try { parsed = JSON.parse(resultJson); }
    catch { setResultInputError("Проверьте синтаксис JSON перед сохранением."); return; }
    if (!parsed || typeof parsed !== "object" || !("entries" in parsed) || !Array.isArray(parsed.entries)) {
      setResultInputError("Передайте объект с массивом entries.");
      return;
    }
    setResultInputError(null);
    void resultAction(() => competitionResultApi.save(competitionId, resultEventId, parsed as Parameters<typeof competitionResultApi.save>[2]), "Результаты сохранены");
  }

  async function downloadProtocol(eventId: number) {
    setPending(true);
    try {
      const file = await competitionResultApi.protocol(competitionId, eventId);
      downloadBlob(file.blob, file.filename);
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось скачать протокол"); }
    finally { setPending(false); }
  }

  if (state.loading) return <DataLoading rows={5} />;
  if (state.error) return <DataError message={state.error} retry={state.reload} />;
  if (!state.data) return null;

  const { competition, events, judges, activeRole, athletes, ageGroups, judgePositions, judgeUsers, canManageCompetition, athleteNames } = state.data;
  const canApply = activeRole === "participant" || activeRole === "coach";
  const applyingEvent = events.find((event) => event.id === applyingEventId);
  const eligibleAthletes = applyingEvent ? athletes.filter((relation) => relation.status === "confirmed" && relation.discipline_id === applyingEvent.discipline.id) : [];
  return (
    <div className="mx-auto w-full max-w-6xl">
      <Button asChild variant="ghost" size="sm" className="mb-4 -ml-2">
        <Link href="/workspace/competitions"><ArrowLeft />К соревнованиям</Link>
      </Button>

      <Card variant="glass" className="p-6">
        <div className="flex flex-col justify-between gap-5 lg:flex-row lg:items-start">
          <div className="max-w-3xl">
            <div className="mb-3 flex flex-wrap gap-2">
              <StatusBadge status={competition.status} />
              <Badge variant="secondary">{competition.discipline.name}</Badge>
              <Badge variant="outline">{competition.competition_type.name}</Badge>
            </div>
            <h2 className="text-2xl font-medium tracking-[-0.02em]">{competition.name}</h2>
            {competition.description ? <p className="mt-2 text-sm leading-6 text-muted-foreground">{competition.description}</p> : null}
          </div>
          <div className="grid shrink-0 gap-2 text-sm text-muted-foreground lg:min-w-80">
            <p className="flex items-center gap-2"><CalendarDays className="size-4" />{formatDate(competition.start_at)} — {formatDate(competition.end_at)}</p>
            {competition.venue_address ? <p className="flex items-center gap-2"><MapPin className="size-4" />{competition.venue_address}</p> : null}
            {competition.regulation_url ? <Button asChild variant="outline" size="sm" className="mt-1 w-fit"><a href={competition.regulation_url} target="_blank" rel="noreferrer"><ExternalLink />Положение</a></Button> : null}
          </div>
        </div>
      </Card>
      {canManageCompetition ? <div className="mt-4 flex flex-wrap gap-2">{competition.status === "draft" ? <Button disabled={pending} onClick={() => transitionCompetition("submit")}>Отправить на проверку</Button> : null}{competition.status === "under_review" ? <><Button disabled={pending} onClick={() => transitionCompetition("publish")}>Опубликовать</Button><Button variant="outline" disabled={pending} onClick={() => transitionCompetition("return")}>Вернуть на доработку</Button></> : null}{competition.status === "under_review" || competition.status === "published" ? <ConfirmAction trigger={<Button variant="outline" disabled={pending}>Отменить соревнование</Button>} title="Отменить соревнование?" description="Соревнование станет недоступно участникам. Вернуть его в публикацию через текущий API нельзя." confirmLabel="Отменить соревнование" destructive onConfirm={() => transitionCompetition("cancel")} /> : null}</div> : null}
      {canUpdate && competition.status === "draft" ? <Card variant="glass" className="mt-4 p-4"><form className="grid gap-2 md:grid-cols-[1fr_1fr_auto]" onSubmit={updateCompetition}><Input name="name" defaultValue={competition.name} required /><Input name="description" defaultValue={competition.description ?? ""} placeholder="Описание" /><Button disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Сохранить карточку</Button></form></Card> : null}

      <div className="mt-5 grid gap-5 lg:grid-cols-[1.4fr_1fr]">
        <Card variant="glass" className="p-5">
          <CardHeader className="flex-row items-start justify-between gap-3 p-0">
            <div>
            <CardTitle>Зачёты соревнования</CardTitle>
            <CardDescription>Дисциплины и возрастные группы.</CardDescription>
            </div>
            {canCreateEvents && competition.status === "draft" ? <Button size="sm" onClick={() => setEventEditor({ mode: "create", ageGroupId: null })}><Plus />Создать</Button> : null}
          </CardHeader>
          <div className="mt-4 grid gap-3">
            {events.map((event) => (
              <div key={event.id} className="flex flex-col gap-4 rounded-xl bg-muted/60 p-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex min-w-0 gap-3">
                  <div>
                    <p className="font-medium">{event.discipline.name}</p>
                    <p className="mt-1 text-sm text-muted-foreground">{event.age_group?.name ?? "Без возрастной группы"}</p>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">{canApply && competition.status === "registration_open" ? <Button size="sm" disabled={pending} onClick={() => setApplyingEventId(event.id)}>Подать заявку</Button> : null}{canReadResults ? <Button variant="outline" size="sm" disabled={pending} onClick={() => void openResults(event.id)}>Результаты</Button> : null}{event.protocol_url ? <Button variant="outline" size="sm" disabled={pending} onClick={() => void downloadProtocol(event.id)}><FileText />Протокол</Button> : null}{canUpdateEvents && competition.status === "draft" ? <Button variant="ghost" size="icon" aria-label={`Изменить зачёт №${event.id}`} onClick={() => setEventEditor({ mode: "edit", eventId: event.id, ageGroupId: event.age_group?.id ?? null })}><Pencil /></Button> : null}{canDeleteEvents && competition.status === "draft" ? <Button variant="ghost" size="icon" className="text-destructive hover:bg-destructive/10 hover:text-destructive" aria-label={`Удалить зачёт №${event.id}`} onClick={() => setDeletingEvent(event)}><Trash2 /></Button> : null}</div>
              </div>
            ))}
            {!events.length ? <p className="text-sm text-muted-foreground">Зачёты пока не опубликованы.</p> : null}
          </div>
        </Card>

        <Card variant="glass" className="p-5">
          <CardHeader>
            <CardTitle>Судейская коллегия</CardTitle>
            <CardDescription>Назначенные судьи соревнования.</CardDescription>
          </CardHeader>
          <div className="mt-4 grid gap-3">
            {canManageJudges && competition.status === "draft" && canReadRegistry ? <form className="grid gap-2 rounded-xl border p-3" onSubmit={assignJudge}><Select name="judge_user_id" required><SelectTrigger><SelectValue placeholder="Выберите судью" /></SelectTrigger><SelectContent>{judgeUsers.map((judge) => <SelectItem key={judge.id} value={String(judge.id)}>{fullName(judge.last_name, judge.first_name, judge.middle_name)}</SelectItem>)}</SelectContent></Select><Select name="position_id" required><SelectTrigger><SelectValue placeholder="Выберите должность" /></SelectTrigger><SelectContent>{judgePositions.map((position) => <SelectItem key={position.id} value={String(position.id)}>{position.name}</SelectItem>)}</SelectContent></Select><Button size="sm" disabled={pending || !judgeUsers.length || !judgePositions.length}>Назначить судью</Button></form> : null}
            {canManageJudges && competition.status === "draft" && !canReadRegistry ? <p className="text-sm text-muted-foreground">Для выбора судьи требуется разрешение на просмотр реестра пользователей.</p> : null}
            {canManageJudges && competition.status !== "draft" ? <p className="text-sm text-muted-foreground">Состав судей можно менять только до публикации соревнования.</p> : null}
            {judges.map((assignment) => (
              <div key={assignment.id} className="flex gap-3 rounded-xl bg-muted/60 p-4">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{fullName(assignment.judge.last_name, assignment.judge.first_name, assignment.judge.middle_name)}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{assignment.position.name}</p>
                </div>
                {canManageJudges && competition.status === "draft" ? <div className="ml-auto"><ConfirmAction trigger={<Button variant="outline" size="sm" disabled={pending}>Снять</Button>} title="Снять судью с назначения?" description="Назначение будет удалено из судейской коллегии этого соревнования." confirmLabel="Снять судью" destructive onConfirm={() => run(() => competitionJudgeApi.remove(competitionId, assignment.id), "Судья снят")} /></div> : null}
              </div>
            ))}
            {!judges.length ? <p className="text-sm text-muted-foreground">Состав пока не опубликован.</p> : null}
          </div>
        </Card>
      </div>
      <Dialog open={Boolean(result && resultEventId)} onOpenChange={(open) => { if (!open && !pending) { setResult(null); setResultEventId(null); setResultInputError(null); } }}>
        <DialogContent className="sm:max-w-4xl">
          {result && resultEventId ? <><DialogHeader><DialogTitle>Результаты: {result.event.discipline_name}</DialogTitle><DialogDescription>Статус {result.status}, задач: {result.task_count}.</DialogDescription></DialogHeader>
        {canManageResults && result.status !== "approved" ? <div className="mt-4 grid gap-3"><form className="grid gap-2 sm:grid-cols-[140px_120px_1fr_auto]" onSubmit={configureResults}><Select name="mode" defaultValue={result.mode}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="individual">Индивидуальный</SelectItem><SelectItem value="team">Командный</SelectItem></SelectContent></Select><Input name="task_count" type="number" min="1" max="50" defaultValue={result.task_count || 1} required /><Input name="teams" defaultValue="[]" aria-label="Составы команд в JSON" placeholder='Команды: [{"team_id":1,"application_ids":[2]}]' /><Button disabled={pending}>Настроить</Button><label className="col-span-full flex items-center gap-2 text-sm"><Checkbox name="confirm_reset" />Сбросить введённые результаты при повторной настройке</label></form>
          <ResultEntryEditor key={`${resultEventId}-${resultVersion}`} table={result} pending={pending} onSave={(payload) => void resultAction(() => competitionResultApi.save(competitionId, resultEventId, payload), "Результат сохранён")} />
          <details className="rounded-xl border p-3"><summary className="cursor-pointer text-sm font-medium">Расширенный ввод JSON</summary><p className="mt-2 text-sm text-muted-foreground">Для массового изменения строк. Статусы: finished, did_not_start, disqualified; задачи: solved, attempted, not_attempted.</p><Textarea className="mt-2 min-h-64 font-mono text-xs" value={resultJson} onChange={(event) => { setResultJson(event.target.value); setResultInputError(null); }} aria-label="JSON результатов" aria-invalid={Boolean(resultInputError)} /><Button className="mt-2" type="button" disabled={pending || !result.entries.length} onClick={saveResultDraft}>Сохранить JSON</Button></details>
          {resultInputError ? <p role="alert" className="text-sm text-destructive">{resultInputError}</p> : null}
          <div className="flex flex-wrap gap-2"><Button variant="outline" disabled={pending || !result.entries.length} onClick={() => void resultAction(() => competitionResultApi.calculate(competitionId, resultEventId), "Места рассчитаны")}>Рассчитать</Button>{canApproveResults ? <Button variant="outline" disabled={pending || !result.entries.length} onClick={() => void resultAction(() => competitionResultApi.approve(competitionId, resultEventId), "Результаты утверждены")}>Утвердить</Button> : null}</div>
        </div> : <div className="mt-4 grid gap-2">{result.entries.map((entry) => <div key={entry.id} className="rounded-xl bg-muted/60 p-3 text-sm">#{entry.id} · место {entry.place ?? "—"} · решено {entry.solved_count}</div>)}</div>}</> : null}
        </DialogContent>
      </Dialog>
      <Dialog open={eventEditor !== null} onOpenChange={(open) => { if (!open && !pending) setEventEditor(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{eventEditor?.mode === "create" ? "Новый зачёт" : "Изменение зачёта"}</DialogTitle>
            <DialogDescription>Выберите возрастную группу для зачёта. Дисциплина берётся из соревнования.</DialogDescription>
          </DialogHeader>
          <form onSubmit={saveEvent} className="grid gap-4">
            <ScrollMask className="grid max-h-72 gap-2 pr-1">
              <Button type="button" variant={eventEditor?.ageGroupId === null ? "default" : "outline"} className="justify-start" onClick={() => setEventEditor((value) => value ? { ...value, ageGroupId: null } : value)}>Без возрастной группы</Button>
              {ageGroups.map((group) => <Button key={group.id} type="button" variant={eventEditor?.ageGroupId === group.id ? "default" : "outline"} className="justify-start" onClick={() => setEventEditor((value) => value ? { ...value, ageGroupId: group.id } : value)}>{group.name}</Button>)}
            </ScrollMask>
            <DialogFooter>
              <Button type="button" variant="outline" disabled={pending} onClick={() => setEventEditor(null)}>Отмена</Button>
              <Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Сохранить</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={deletingEvent !== null} onOpenChange={(open) => { if (!open && !pending) setDeletingEvent(null); }}>
        <DialogContent>
          <DialogHeader><DialogTitle>Удалить зачёт?</DialogTitle><DialogDescription>Удаление доступно только для черновика без заявок. Отменить его нельзя.</DialogDescription></DialogHeader>
          <DialogFooter>
            <Button variant="outline" disabled={pending} onClick={() => setDeletingEvent(null)}>Отмена</Button>
            <Button variant="destructive" disabled={pending} onClick={removeEvent}>{pending ? <LoaderCircle className="animate-spin" /> : <Trash2 />}Удалить</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={applyingEventId !== null} onOpenChange={(open) => { if (!open) setApplyingEventId(null); }}>
        <DialogContent><DialogHeader><DialogTitle>Заявка на соревнование</DialogTitle><DialogDescription>{applyingEvent?.discipline.name}{applyingEvent?.age_group ? ` · ${applyingEvent.age_group.name}` : ""}. Заявка будет отправлена на рассмотрение федерации.</DialogDescription></DialogHeader>
          {activeRole === "coach" && !eligibleAthletes.length ? <div className="grid gap-3 text-sm text-muted-foreground"><p>Нет подтверждённых спортсменов по этой дисциплине. Сначала создайте связь и дождитесь подтверждения спортсмена.</p><Button asChild variant="outline"><Link href="/workspace/coaches">К моим спортсменам</Link></Button></div> : <form className="grid gap-4" onSubmit={applyForEvent}>
            {activeRole === "coach" ? <><label htmlFor="application-athlete" className="text-sm font-medium">Спортсмен</label><Select name="athlete_user_id" required><SelectTrigger id="application-athlete"><SelectValue placeholder="Выберите спортсмена" /></SelectTrigger><SelectContent>{eligibleAthletes.map((relation) => <SelectItem key={relation.participant_user_id} value={String(relation.participant_user_id)}>{athleteNames[relation.participant_user_id] ?? "Спортсмен"}</SelectItem>)}</SelectContent></Select></> : <p className="text-sm text-muted-foreground">Заявка будет подана за вас.</p>}
            <DialogFooter><Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Отправить заявку</Button></DialogFooter>
          </form>}
        </DialogContent>
      </Dialog>
    </div>
  );
}
