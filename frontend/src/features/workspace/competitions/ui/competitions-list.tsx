"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import { CalendarDays, LoaderCircle, MapPin, Plus, Search, Video } from "lucide-react";
import Link from "next/link";

import { competitionApi } from "@/entities/competition";
import { referenceDataApi } from "@/entities/reference-data";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { useNotifications } from "@/shared/lib/notifications";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";

const selectClassName = "h-9 w-full rounded-[10px] border border-border bg-background px-2.5 text-sm outline-none";

export function CompetitionsList() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const canCreate = identity?.permissions?.includes("competitions.create") ?? false;
  const [search, setSearch] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const load = useCallback(async () => {
    const now = new Date();
    const from = new Date(now.getFullYear() - 1, 0, 1).toISOString();
    const to = new Date(now.getFullYear() + 2, 0, 1).toISOString();
    const [list, calendar, disciplines, types, levels] = await Promise.all([
      competitionApi.list({ limit: 100 }),
      competitionApi.calendar({ from, to, limit: 100 }),
      referenceDataApi.disciplines(),
      referenceDataApi.competitionTypes(),
      referenceDataApi.competitionLevels(),
    ]);
    return { list, calendar, disciplines, types, levels };
  }, []);
  const state = useAsyncData(load);
  const competitions = useMemo(() => {
    const query = search.trim().toLocaleLowerCase("ru");
    return (state.data?.list ?? []).filter((competition) =>
      !query
        ? true
        : [competition.name, competition.discipline.name, competition.competition_type.name]
            .join(" ")
            .toLocaleLowerCase("ru")
            .includes(query),
    );
  }, [search, state.data]);

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const nullable = (name: string) => String(form.get(name) ?? "").trim() || null;
    setPending(true);
    try {
      await competitionApi.create({
        name: String(form.get("name") ?? "").trim(),
        description: nullable("description"),
        discipline_id: Number(form.get("discipline_id")),
        competition_type_id: Number(form.get("competition_type_id")),
        competition_level_id: Number(form.get("competition_level_id")),
        format: String(form.get("format")),
        start_at: new Date(String(form.get("start_at"))).toISOString(),
        end_at: new Date(String(form.get("end_at"))).toISOString(),
        registration_starts_at: new Date(String(form.get("registration_starts_at"))).toISOString(),
        registration_ends_at: new Date(String(form.get("registration_ends_at"))).toISOString(),
        venue_address: nullable("venue_address"),
        online_url: nullable("online_url"),
      });
      notify.success("Соревнование создано");
      setCreateOpen(false);
      state.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось создать соревнование"); }
    finally { setPending(false); }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading
        title="Соревнования"
        description="Опубликованные соревнования и сроки регистрации."
        action={canCreate ? <Button onClick={() => setCreateOpen(true)}><Plus />Создать</Button> : undefined}
      />
      {state.data ? <p className="mb-3 text-xs text-muted-foreground">В календаре на текущий период: {state.data.calendar.length}</p> : null}
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

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl"><DialogHeader><DialogTitle>Новое соревнование</DialogTitle><DialogDescription>Заполните основные параметры. Новый объект будет создан как черновик.</DialogDescription></DialogHeader>
          <form className="grid gap-3" onSubmit={create}>
            <Label htmlFor="competition-name">Название</Label><Input id="competition-name" name="name" required />
            <Label htmlFor="competition-description">Описание</Label><Input id="competition-description" name="description" />
            <div className="grid gap-3 sm:grid-cols-3">
              <select name="discipline_id" className={selectClassName} required defaultValue=""><option value="" disabled>Дисциплина</option>{state.data?.disciplines.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select>
              <select name="competition_type_id" className={selectClassName} required defaultValue=""><option value="" disabled>Тип</option>{state.data?.types.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select>
              <select name="competition_level_id" className={selectClassName} required defaultValue=""><option value="" disabled>Уровень</option>{state.data?.levels.map((item) => <option key={item.id} value={item.id}>{item.name ?? item.code ?? `#${item.id}`}</option>)}</select>
            </div>
            <Label htmlFor="competition-format">Формат</Label><select id="competition-format" name="format" className={selectClassName} defaultValue="offline"><option value="offline">Очно</option><option value="online">Онлайн</option><option value="hybrid">Гибрид</option></select>
            <div className="grid gap-3 sm:grid-cols-2"><div><Label htmlFor="start-at">Начало</Label><Input id="start-at" name="start_at" type="datetime-local" required /></div><div><Label htmlFor="end-at">Окончание</Label><Input id="end-at" name="end_at" type="datetime-local" required /></div><div><Label htmlFor="registration-starts-at">Начало регистрации</Label><Input id="registration-starts-at" name="registration_starts_at" type="datetime-local" required /></div><div><Label htmlFor="registration-ends-at">Конец регистрации</Label><Input id="registration-ends-at" name="registration_ends_at" type="datetime-local" required /></div></div>
            <Input name="venue_address" placeholder="Адрес площадки" /><Input name="online_url" type="url" placeholder="Ссылка на онлайн-площадку" />
            <DialogFooter><Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Создать</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
