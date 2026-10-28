"use client";

import { FormEvent, useCallback, useState } from "react";
import {
  ArrowLeft,
  CalendarDays,
  ExternalLink,
  FileText,
  Gavel,
  LoaderCircle,
  MapPin,
  Trophy,
} from "lucide-react";
import Link from "next/link";

import { competitionApi } from "@/entities/competition";
import { competitionApplicationApi } from "@/entities/competition-application";
import { competitionEventApi } from "@/entities/competition-event";
import { competitionJudgeApi } from "@/entities/competition-judge";
import { competitionResultApi, type CompetitionResultTable } from "@/entities/competition-result";
import { profileApi } from "@/entities/profile";
import { useAuth } from "@/entities/session";
import { userCoachApi } from "@/entities/user-coach";
import { formatDate, fullName, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { DataError, DataLoading } from "@/shared/ui/data-state";
import { useNotifications } from "@/shared/lib/notifications";
import { Input } from "@/shared/ui/input";

export function CompetitionDetails({ competitionId }: { competitionId: number }) {
  const { identity } = useAuth();
  const notify = useNotifications();
  const permissions = new Set(identity?.permissions ?? []);
  const canUpdate = permissions.has("competitions.update");
  const canManageJudges = permissions.has("competition_judges.manage");
  const canReadResults = permissions.has("competition_results.read");
  const canManageResults = permissions.has("competition_results.manage");
  const canApproveResults = permissions.has("competition_results.approve");
  const [pending, setPending] = useState(false);
  const [result, setResult] = useState<CompetitionResultTable | null>(null);
  const [resultEventId, setResultEventId] = useState<number | null>(null);
  const [resultJson, setResultJson] = useState("");
  const [athleteByEvent, setAthleteByEvent] = useState<Record<number, string>>({});
  const load = useCallback(
    async () => {
      const profile = await profileApi.me();
      const activeRole = profile.roles.find((role) => role.id === identity?.roleId)?.code ?? null;
      const [competition, events, judges, applications, coachRelations] = await Promise.all([
        competitionApi.get(competitionId),
        competitionEventApi.publicList(competitionId),
        canManageJudges ? competitionJudgeApi.adminList(competitionId) : competitionJudgeApi.publicList(competitionId),
        competitionApplicationApi.mine(),
        activeRole === "coach" ? userCoachApi.byCoach({ limit: 100 }) : Promise.resolve([]),
      ]);
      const now = Date.now();
      const registrationOpen = competition.status === "registration_open" &&
        now >= new Date(competition.registration_starts_at).getTime() &&
        now <= new Date(competition.registration_ends_at).getTime();
      return { competition, events, judges, applications: applications.filter((item) => item.competition_id === competitionId), coachRelations, activeRole, registrationOpen };
    },
    [canManageJudges, competitionId, identity?.roleId],
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

  function assignJudge(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(() => competitionJudgeApi.assign(competitionId, { judge_user_id: Number(form.get("judge_user_id")), position_id: Number(form.get("position_id")) }), "Судья назначен");
  }

  function submitApplication(eventId: number) {
    const selected = Number(athleteByEvent[eventId]);
    void run(
      () => competitionApplicationApi.create({ event_id: eventId, athlete_user_id: selected > 0 ? selected : undefined }),
      "Заявка на соревнование подана",
    );
  }

  async function openResults(eventId: number) {
    setPending(true);
    try {
      if (permissions.has("competition_events.read")) await competitionEventApi.adminGet(competitionId, eventId);
      else await competitionEventApi.publicGet(competitionId, eventId);
      const table = await competitionResultApi.get(competitionId, eventId);
      setResult(table);
      setResultEventId(eventId);
      setResultJson(JSON.stringify({ entries: table.entries.map((entry) => ({ entry_id: entry.id, result_status: entry.result_status, tasks: entry.tasks.map((task) => ({ task_number: task.task_number, status: task.status, solve_time_deciseconds: task.solve_time_deciseconds, negative_attempts: task.negative_attempts })) })) }, null, 2));
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить результаты"); }
    finally { setPending(false); }
  }

  async function resultAction(action: () => Promise<CompetitionResultTable | unknown>, success: string) {
    setPending(true);
    try {
      const next = await action();
      if (next && typeof next === "object" && "entries" in next) setResult(next as CompetitionResultTable);
      notify.success(success);
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось обновить результаты"); }
    finally { setPending(false); }
  }

  async function downloadProtocol(eventId: number) {
    setPending(true);
    try {
      const file = await competitionResultApi.protocol(competitionId, eventId);
      const url = URL.createObjectURL(file.blob);
      const anchor = document.createElement("a"); anchor.href = url; anchor.download = file.filename; anchor.click(); URL.revokeObjectURL(url);
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось скачать протокол"); }
    finally { setPending(false); }
  }

  if (state.loading) return <DataLoading rows={5} />;
  if (state.error) return <DataError message={state.error} retry={state.reload} />;
  if (!state.data) return null;

  const { competition, events, judges, applications, coachRelations, activeRole, registrationOpen } = state.data;
  return (
    <div className="mx-auto w-full max-w-6xl">
      <Button asChild variant="ghost" size="sm" className="mb-4 -ml-2">
        <Link href="/workspace/competitions"><ArrowLeft />К соревнованиям</Link>
      </Button>

      <Card className="bg-white/75 p-6 shadow-none">
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
      {canUpdate ? <Card className="mt-4 bg-white/75 p-4 shadow-none"><form className="grid gap-2 md:grid-cols-[1fr_1fr_auto]" onSubmit={updateCompetition}><Input name="name" defaultValue={competition.name} required /><Input name="description" defaultValue={competition.description ?? ""} placeholder="Описание" /><Button disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Сохранить карточку</Button></form></Card> : null}

      <div className="mt-5 grid gap-5 lg:grid-cols-[1.4fr_1fr]">
        <Card className="bg-white/75 p-5 shadow-none">
          <CardHeader>
            <CardTitle>Зачёты соревнования</CardTitle>
            <CardDescription>Дисциплины и возрастные группы.</CardDescription>
          </CardHeader>
          <div className="mt-4 grid gap-3">
            {events.map((event) => (
              <div key={event.id} className="flex items-center justify-between gap-4 rounded-xl bg-muted/60 p-4">
                <div className="flex min-w-0 gap-3">
                  <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Trophy className="size-5" /></span>
                  <div>
                    <p className="font-medium">{event.discipline.name}</p>
                    <p className="mt-1 text-sm text-muted-foreground">{event.age_group?.name ?? "Без возрастной группы"}</p>
                  </div>
                </div>
                <div className="flex flex-wrap justify-end gap-2">
                  {activeRole === "coach" && registrationOpen ? <select value={athleteByEvent[event.id] ?? ""} onChange={(change) => setAthleteByEvent((current) => ({ ...current, [event.id]: change.target.value }))} className="h-8 rounded-[10px] border border-border bg-background px-2 text-xs"><option value="">Выберите спортсмена</option>{coachRelations.filter((relation) => relation.status === "confirmed" && relation.discipline_id === event.discipline.id).map((relation) => <option key={relation.id} value={relation.participant_user_id}>Спортсмен #{relation.participant_user_id}</option>)}</select> : null}
                  {registrationOpen && !applications.some((application) => application.event_id === event.id && (activeRole !== "coach" || application.athlete_user_id === Number(athleteByEvent[event.id]))) ? <Button size="sm" disabled={pending || (activeRole === "coach" && !athleteByEvent[event.id])} onClick={() => submitApplication(event.id)}>Подать заявку</Button> : null}
                  {applications.filter((application) => application.event_id === event.id && (activeRole !== "coach" || !athleteByEvent[event.id] || application.athlete_user_id === Number(athleteByEvent[event.id]))).map((application) => <StatusBadge key={application.id} status={application.status} />)}
                  {canReadResults ? <Button variant="outline" size="sm" disabled={pending} onClick={() => void openResults(event.id)}>Результаты</Button> : null}{event.protocol_url ? <Button variant="outline" size="sm" disabled={pending} onClick={() => void downloadProtocol(event.id)}><FileText />Протокол</Button> : null}
                </div>
              </div>
            ))}
            {!events.length ? <p className="text-sm text-muted-foreground">Зачёты пока не опубликованы.</p> : null}
          </div>
        </Card>

        <Card className="bg-white/75 p-5 shadow-none">
          <CardHeader>
            <CardTitle>Судейская коллегия</CardTitle>
            <CardDescription>Назначенные судьи соревнования.</CardDescription>
          </CardHeader>
          <div className="mt-4 grid gap-3">
            {canManageJudges ? <form className="grid gap-2 rounded-xl border p-3" onSubmit={assignJudge}><Input name="judge_user_id" type="number" min="1" placeholder="ID судьи" required /><Input name="position_id" type="number" min="1" placeholder="ID позиции" required /><Button size="sm" disabled={pending}>Назначить судью</Button></form> : null}
            {judges.map((assignment) => (
              <div key={assignment.id} className="flex gap-3 rounded-xl bg-muted/60 p-4">
                <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-white text-primary"><Gavel className="size-4" /></span>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{fullName(assignment.judge.last_name, assignment.judge.first_name, assignment.judge.middle_name)}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{assignment.position.name}</p>
                </div>
                {canManageJudges ? <Button className="ml-auto" variant="outline" size="sm" disabled={pending} onClick={() => void run(() => competitionJudgeApi.remove(competitionId, assignment.id), "Судья снят")}>Снять</Button> : null}
              </div>
            ))}
            {!judges.length ? <p className="text-sm text-muted-foreground">Состав пока не опубликован.</p> : null}
          </div>
        </Card>
      </div>
      {result && resultEventId ? <Card className="mt-5 bg-white/75 p-5 shadow-none"><CardHeader><CardTitle>Результаты: {result.event.discipline_name}</CardTitle><CardDescription>Статус {result.status}, задач: {result.task_count}.</CardDescription></CardHeader>
        {canManageResults ? <div className="mt-4 grid gap-3"><form className="grid gap-2 sm:grid-cols-[140px_120px_1fr_auto]" onSubmit={(event) => { event.preventDefault(); const form = new FormData(event.currentTarget); void resultAction(() => competitionResultApi.configure(competitionId, resultEventId, { mode: String(form.get("mode")), task_count: Number(form.get("task_count")), teams: JSON.parse(String(form.get("teams") || "[]")), confirm_reset: true }), "Настройки сохранены"); }}><select name="mode" defaultValue={result.mode} className="h-9 rounded-[10px] border border-border bg-background px-2.5 text-sm"><option value="individual">Индивидуальный</option><option value="team">Командный</option></select><Input name="task_count" type="number" min="1" defaultValue={result.task_count || 1} /><Input name="teams" defaultValue="[]" placeholder='Команды: [{"team_id":1,"application_ids":[2]}]' /><Button disabled={pending}>Настроить</Button></form>
          <textarea className="min-h-64 rounded-xl border bg-background p-3 font-mono text-xs" value={resultJson} onChange={(event) => setResultJson(event.target.value)} aria-label="JSON результатов" />
          <div className="flex flex-wrap gap-2"><Button disabled={pending} onClick={() => void resultAction(() => competitionResultApi.save(competitionId, resultEventId, JSON.parse(resultJson)), "Результаты сохранены")}>Сохранить таблицу</Button><Button variant="outline" disabled={pending} onClick={() => void resultAction(() => competitionResultApi.calculate(competitionId, resultEventId), "Места рассчитаны")}>Рассчитать</Button>{canApproveResults ? <Button variant="outline" disabled={pending} onClick={() => void resultAction(() => competitionResultApi.approve(competitionId, resultEventId), "Результаты утверждены")}>Утвердить</Button> : null}</div>
        </div> : <div className="mt-4 grid gap-2">{result.entries.map((entry) => <div key={entry.id} className="rounded-xl bg-muted/60 p-3 text-sm">#{entry.id} · место {entry.place ?? "—"} · решено {entry.solved_count}</div>)}</div>}
      </Card> : null}
    </div>
  );
}
