"use client";

import { useCallback, useMemo, useState } from "react";
import Link from "next/link";
import { AlertTriangle, Download, FileText, LoaderCircle } from "lucide-react";

import { competitionResultApi } from "@/entities/competition-result";
import { profileApi, type PublicProfile } from "@/entities/profile";
import { referenceDataApi, type Discipline } from "@/entities/reference-data";
import { useAuth } from "@/entities/session";
import { userCoachApi, type UserCoachRelation } from "@/entities/user-coach";
import {
  expiryDescription,
  formatDate,
  fullName,
  isExpiringSoon,
  StatusBadge,
} from "@/features/workspace/common";
import { downloadBlob } from "@/shared/lib/download";
import { useNotifications } from "@/shared/lib/notifications";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { ChartLine } from "@/shared/ui/chart";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";

interface ResultsContext {
  role: string;
  relations: UserCoachRelation[];
  names: Record<number, string>;
  profiles: Record<number, PublicProfile>;
  disciplines: Discipline[];
}

export function ResultsHistory() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const [selectedRelationId, setSelectedRelationId] = useState(0);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const [downloading, setDownloading] = useState(false);

  const context = useAsyncData(useCallback(async (): Promise<ResultsContext> => {
    const profile = await profileApi.me();
    const role = profile.roles.find((item) => item.id === identity?.roleId)?.code ?? "participant";
    if (role !== "coach") return { role, relations: [], names: {}, profiles: {}, disciplines: [] };
    const [relations, disciplines] = await Promise.all([
      userCoachApi.byCoach({ limit: 100 }).then((items) => items.filter((item) => item.status === "confirmed")),
      referenceDataApi.disciplines(),
    ]);
    const userIds = [...new Set(relations.map((item) => item.participant_user_id))];
    const profiles = await Promise.allSettled(userIds.map((userId) => profileApi.public(userId)));
    const names = Object.fromEntries(profiles.flatMap((item, index) => item.status === "fulfilled" ? [[userIds[index], fullName(item.value.last_name, item.value.first_name, item.value.middle_name)]] : []));
    const profileMap = Object.fromEntries(profiles.flatMap((item, index) => item.status === "fulfilled" ? [[userIds[index], item.value]] : []));
    return { role, relations, names, profiles: profileMap, disciplines };
  }, [identity?.roleId]));

  const selectedRelation = context.data?.relations.find((item) => item.id === selectedRelationId) ?? context.data?.relations[0] ?? null;
  const contextRole = context.data?.role;
  const selectedAthleteUserId = selectedRelation?.participant_user_id;
  const selectedDisciplineId = selectedRelation?.discipline_id;
  const selectedAthleteProfile = selectedAthleteUserId
    ? context.data?.profiles[selectedAthleteUserId]
    : undefined;
  const expiringRanks = (selectedAthleteProfile?.ranks ?? []).filter(
    (rank) =>
      rank.discipline_id === selectedDisciplineId &&
      isExpiringSoon(rank.expires_at),
  );
  const results = useAsyncData(useCallback(() => {
    if (!contextRole) return Promise.resolve([]);
    if (contextRole === "coach") {
      if (!selectedAthleteUserId || !selectedDisciplineId) return Promise.resolve([]);
      return competitionResultApi.history({ athlete_user_id: selectedAthleteUserId, discipline_id: selectedDisciplineId });
    }
    return competitionResultApi.history();
  }, [contextRole, selectedAthleteUserId, selectedDisciplineId]));

  const items = useMemo(() => (results.data ?? []).filter((item) => `${item.competition_name} ${item.discipline_name}`.toLocaleLowerCase("ru").includes(search.trim().toLocaleLowerCase("ru")) && (status === "all" || item.result_status === status)), [search, results.data, status]);
  const chartItems = [...items].filter((item) => item.score != null || item.place != null).sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime());
  const scorePoints = chartItems.filter((item) => item.score != null).map((item) => ({ label: formatDate(item.start_at), value: item.score as number }));
  const placePoints = chartItems.filter((item) => item.place != null).map((item) => ({ label: formatDate(item.start_at), value: item.place as number }));
  const loading = context.loading || results.loading;
  const error = context.error ?? results.error;

  async function downloadReport() {
    setDownloading(true);
    try {
      const report = await profileApi.report();
      downloadBlob(report.blob, report.filename);
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось сформировать отчёт");
    } finally {
      setDownloading(false);
    }
  }

  const resultContext = context.data;

  return (
    <div className="mx-auto w-full max-w-6xl">
      {resultContext?.role === "coach" ? (
        <Card variant="glass" className="mb-4 p-4 sm:p-5">
          <CardHeader className="p-0"><CardTitle>Результаты спортсмена</CardTitle><CardDescription>Выберите подтверждённую связь. История ограничена дисциплиной, по которой спортсмен подтвердил тренера.</CardDescription></CardHeader>
          {resultContext.relations.length ? <Select value={selectedRelation ? String(selectedRelation.id) : ""} onValueChange={(value) => setSelectedRelationId(Number(value))}><SelectTrigger className="mt-4"><SelectValue placeholder="Выберите спортсмена" /></SelectTrigger><SelectContent>{resultContext.relations.map((relation) => <SelectItem key={relation.id} value={String(relation.id)}>{resultContext.names[relation.participant_user_id] ?? "Спортсмен"} · {resultContext.disciplines.find((item) => item.id === relation.discipline_id)?.name ?? "Дисциплина"}</SelectItem>)}</SelectContent></Select> : <div className="mt-4"><DataEmpty title="Нет подтверждённых спортсменов" description="Сначала отправьте запрос спортсмену и дождитесь подтверждения связи." action={<Button asChild variant="outline" size="sm"><Link href="/workspace/coaches">Открыть спортсменов</Link></Button>} /></div>}
        </Card>
      ) : <div className="mb-3 flex justify-end"><Button type="button" variant="outline" disabled={downloading} onClick={() => void downloadReport()}>{downloading ? <LoaderCircle className="animate-spin" /> : <Download />}Скачать отчёт по спортсмену</Button></div>}

      {resultContext?.role === "coach" && expiringRanks.length ? (
        <Alert className="mb-4 border-amber-500/30 bg-amber-500/10 text-foreground">
          <AlertTriangle className="text-amber-600" />
          <AlertTitle>У спортсмена скоро истекает разряд</AlertTitle>
          <AlertDescription className="grid gap-2">
            {expiringRanks.map((rank) => (
              <p key={rank.id}>
                {rank.rank_name} · {rank.discipline_name}: {expiryDescription(rank.expires_at)} ({formatDate(rank.expires_at)}).
              </p>
            ))}
            <Button asChild variant="outline" size="sm" className="mt-1 w-fit bg-background">
              <Link href="/workspace/applications">Подать заявку за спортсмена</Link>
            </Button>
          </AlertDescription>
        </Alert>
      ) : null}

      <WorkspaceToolbar search={search} onSearchChange={setSearch} searchPlaceholder="Найти результат" filterActive={status !== "all"} onResetFilters={() => setStatus("all")} filterContent={<Select value={status} onValueChange={setStatus}><SelectTrigger><SelectValue placeholder="Статус" /></SelectTrigger><SelectContent><SelectItem value="all">Все статусы</SelectItem><SelectItem value="approved">Утверждённые</SelectItem><SelectItem value="finished">Завершённые</SelectItem></SelectContent></Select>} />
      {loading ? <DataLoading rows={4} /> : null}
      {error ? <DataError message={error} retry={() => { context.reload(); results.reload(); }} /> : null}
      {!loading && !error && !(resultContext?.role === "coach" && !selectedRelation) && !items.length ? <DataEmpty title={search || status !== "all" ? "Результаты не найдены" : "Результатов пока нет"} description={search || status !== "all" ? "Измените запрос или сбросьте фильтры." : "После утверждения результатов соревнований они появятся здесь."} action={search || status !== "all" ? <Button variant="outline" size="sm" onClick={() => { setSearch(""); setStatus("all"); }}>Сбросить фильтры</Button> : undefined} /> : null}
      {!loading && !error && (scorePoints.length >= 2 || placePoints.length >= 2) ? <Card variant="glass" className="mb-4 p-5"><CardHeader className="p-0"><CardTitle>Динамика результатов</CardTitle><CardDescription>{scorePoints.length >= 2 ? "Баллы по утверждённым соревнованиям." : "Места по утверждённым соревнованиям; меньшее место лучше."}</CardDescription></CardHeader><div className="mt-4"><ChartLine points={scorePoints.length >= 2 ? scorePoints : placePoints} label={scorePoints.length >= 2 ? "График изменения баллов по соревнованиям" : "График мест по соревнованиям"} /></div></Card> : null}
      <div className="grid gap-3">{items.map((item) => <Card key={`${item.competition_id}-${item.event_id}`} variant="glass" className="p-5"><div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center"><div className="min-w-0"><p className="font-medium">{item.competition_name}</p><p className="mt-1 text-sm text-muted-foreground">{item.discipline_name} · {formatDate(item.start_at)}</p><div className="mt-2 flex flex-wrap items-center gap-2">{item.place ? <span className="text-sm font-medium">{item.place} место</span> : null}{item.score != null ? <span className="text-sm text-muted-foreground">{item.score} баллов</span> : null}<StatusBadge status={item.result_status} /></div></div>{item.protocol_url ? <Button asChild variant="outline" size="sm"><a href={item.protocol_url} target="_blank" rel="noreferrer"><FileText />Протокол</a></Button> : null}</div></Card>)}</div>
    </div>
  );
}
