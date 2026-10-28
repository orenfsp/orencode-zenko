"use client";

import { useCallback, useMemo, useState } from "react";
import { CalendarDays, MapPin, Search, Video } from "lucide-react";
import Link from "next/link";

import { competitionApi } from "@/entities/competition";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";

export function CompetitionsList() {
  const [search, setSearch] = useState("");
  const load = useCallback(() => competitionApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);
  const competitions = useMemo(() => {
    const query = search.trim().toLocaleLowerCase("ru");
    return (state.data ?? []).filter((competition) =>
      !query
        ? true
        : [competition.name, competition.discipline.name, competition.competition_type.name]
            .join(" ")
            .toLocaleLowerCase("ru")
            .includes(query),
    );
  }, [search, state.data]);

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading
        title="Соревнования"
        description="Опубликованные соревнования и сроки регистрации."
      />
      <div className="relative mb-4 max-w-md">
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Найти соревнование"
          className="bg-white/75 pl-9 shadow-none"
        />
      </div>

      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !competitions.length ? (
        <DataEmpty
          title={search ? "Ничего не найдено" : "Соревнований пока нет"}
          description={search ? "Попробуйте изменить запрос." : "Опубликованные мероприятия появятся здесь."}
        />
      ) : null}

      <div className="grid gap-3">
        {competitions.map((competition) => (
          <Card key={competition.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 md:flex-row md:items-start">
              <div className="min-w-0">
                <div className="mb-2 flex flex-wrap items-center gap-2">
                  <StatusBadge status={competition.status} />
                  <span className="text-xs text-muted-foreground">{competition.discipline.name}</span>
                </div>
                <h3 className="text-base font-medium">{competition.name}</h3>
                {competition.description ? (
                  <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">{competition.description}</p>
                ) : null}
              </div>
              <div className="grid shrink-0 gap-2 text-sm text-muted-foreground md:min-w-72">
                <p className="flex items-center gap-2"><CalendarDays className="size-4" />{formatDate(competition.start_at)} — {formatDate(competition.end_at)}</p>
                {competition.venue_address ? <p className="flex items-center gap-2"><MapPin className="size-4" />{competition.venue_address}</p> : null}
                {competition.online_url ? <p className="flex items-center gap-2"><Video className="size-4" />Онлайн-формат</p> : null}
                <Button asChild variant="outline" size="sm" className="mt-1 w-fit">
                  <Link href={`/workspace/competitions/${competition.id}`}>Подробнее</Link>
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
