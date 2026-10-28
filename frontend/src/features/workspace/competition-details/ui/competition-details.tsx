"use client";

import { useCallback } from "react";
import {
  ArrowLeft,
  CalendarDays,
  ExternalLink,
  FileText,
  Gavel,
  MapPin,
  Trophy,
} from "lucide-react";
import Link from "next/link";

import { competitionApi } from "@/entities/competition";
import { competitionEventApi } from "@/entities/competition-event";
import { competitionJudgeApi } from "@/entities/competition-judge";
import { formatDate, fullName, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { DataError, DataLoading } from "@/shared/ui/data-state";

export function CompetitionDetails({ competitionId }: { competitionId: number }) {
  const load = useCallback(
    async () => {
      const [competition, events, judges] = await Promise.all([
        competitionApi.get(competitionId),
        competitionEventApi.publicList(competitionId),
        competitionJudgeApi.publicList(competitionId),
      ]);
      return { competition, events, judges };
    },
    [competitionId],
  );
  const state = useAsyncData(load);

  if (state.loading) return <DataLoading rows={5} />;
  if (state.error) return <DataError message={state.error} retry={state.reload} />;
  if (!state.data) return null;

  const { competition, events, judges } = state.data;
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
                {event.protocol_url ? <Button asChild variant="outline" size="sm"><a href={event.protocol_url} target="_blank" rel="noreferrer"><FileText />Протокол</a></Button> : null}
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
            {judges.map((assignment) => (
              <div key={assignment.id} className="flex gap-3 rounded-xl bg-muted/60 p-4">
                <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-white text-primary"><Gavel className="size-4" /></span>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{fullName(assignment.judge.last_name, assignment.judge.first_name, assignment.judge.middle_name)}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{assignment.position.name}</p>
                </div>
              </div>
            ))}
            {!judges.length ? <p className="text-sm text-muted-foreground">Состав пока не опубликован.</p> : null}
          </div>
        </Card>
      </div>
    </div>
  );
}
