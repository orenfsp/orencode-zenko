"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { CalendarDays, LoaderCircle, MapPin, Plus, Search, Video } from "lucide-react";
import Link from "next/link";

import { competitionApi } from "@/entities/competition";
import type { Competition } from "@/entities/competition/model/types";
import { referenceDataApi } from "@/entities/reference-data";
import { useAuth } from "@/entities/session";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { DateTimePicker } from "@/shared/ui/date-time-picker";
import { Label } from "@/shared/ui/label";
import { useNotifications } from "@/shared/lib/notifications";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";

export function CompetitionsList() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const canCreate = identity?.permissions?.includes("competitions.create") ?? false;
  const [search, setSearch] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [format, setFormat] = useState("offline");
  const [pending, setPending] = useState(false);
  const [recentDrafts, setRecentDrafts] = useState<Competition[]>([]);
  const draftStorageKey = identity ? `workspace-competition-drafts-${identity.userId}` : null;
  useEffect(() => {
    if (!draftStorageKey) return;
    let active = true;
    let drafts: Competition[] = [];
    try {
      const stored: unknown = JSON.parse(sessionStorage.getItem(draftStorageKey) ?? "[]");
      if (Array.isArray(stored)) {
        drafts = stored.filter((item): item is Competition =>
          typeof item === "object" && item !== null && typeof item.id === "number" && item.status === "draft" && typeof item.name === "string",
        );
      }
    } catch { /* Invalid or unavailable tab storage. */ }
    queueMicrotask(() => {
      if (active) setRecentDrafts(drafts);
    });
    return () => { active = false; };
  }, [draftStorageKey]);
  const loadList = useCallback(() => competitionApi.list({ limit: 100 }), []);
  const loadReferences = useCallback(async () => {
    const [disciplines, types, levels] = await Promise.all([
      referenceDataApi.disciplines(),
      referenceDataApi.competitionTypes(),
      referenceDataApi.competitionLevels(),
    ]);
    return { disciplines, types, levels };
  }, []);
  const state = useAsyncData(loadList);
  const references = useAsyncData(loadReferences);
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
  const visibleDrafts = recentDrafts.filter((competition) =>
    competition.name.toLocaleLowerCase("ru").includes(search.trim().toLocaleLowerCase("ru")),
  );

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const nullable = (name: string) => String(form.get(name) ?? "").trim() || null;
    const timestamps = ["start_at", "end_at", "registration_starts_at", "registration_ends_at"].map((name) => new Date(String(form.get(name))).getTime());
    if (timestamps.some((value) => !Number.isFinite(value))) {
      notify.error("Укажите даты соревнования и регистрации");
      return;
    }
    if (timestamps[0] >= timestamps[1] || timestamps[2] >= timestamps[3] || timestamps[3] > timestamps[0]) {
      notify.error("Проверьте даты: регистрация должна закончиться до начала соревнования, начало — быть раньше окончания");
      return;
    }
    setPending(true);
    try {
      const created = await competitionApi.create({
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
      if (draftStorageKey) {
        const drafts = [created, ...recentDrafts.filter((draft) => draft.id !== created.id)];
        setRecentDrafts(drafts);
        try {
          sessionStorage.setItem(draftStorageKey, JSON.stringify(drafts));
        } catch { /* Storage may be unavailable. */ }
      }
      notify.success("Черновик соревнования создан");
      setCreateOpen(false);
      state.reload();
    } catch (cause) { notify.error(cause instanceof Error ? cause.message : "Не удалось создать соревнование"); }
    finally { setPending(false); }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      {canCreate ? <div className="mb-4 flex justify-end"><Button onClick={() => setCreateOpen(true)}><Plus />Создать</Button></div> : null}
      {visibleDrafts.length ? (
        <section className="mb-6" aria-label="Черновики, созданные в этой вкладке">
          <h2 className="mb-2 text-sm font-medium">Созданные черновики</h2>
          <div className="grid gap-3">
            {visibleDrafts.map((draft) => (
              <Card key={draft.id} className="bg-primary/5 p-4 shadow-none">
                <div className="flex flex-wrap items-center gap-2">
                  <StatusBadge status={draft.status} />
                  <span className="text-xs text-muted-foreground">№{draft.id}</span>
                </div>
                <p className="font-medium">{draft.name}</p>
                <p className="text-sm text-muted-foreground">{formatDate(draft.start_at)} — {formatDate(draft.end_at)}</p>
              </Card>
            ))}
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            Это записи, созданные в текущей вкладке. API пока не предоставляет список всех черновиков и их просмотр после повторного входа.
          </p>
        </section>
      ) : null}
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
      {!state.loading && !state.error && !competitions.length && !visibleDrafts.length ? (
        <DataEmpty
          title={search ? "Ничего не найдено" : "Соревнований пока нет"}
          description={search ? "Попробуйте изменить запрос." : "Публичный API показывает только опубликованные соревнования. Новые черновики появятся выше после создания в этой вкладке."}
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
              <Select name="discipline_id"><SelectTrigger><SelectValue placeholder="Дисциплина" /></SelectTrigger><SelectContent>{references.data?.disciplines.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
              <Select name="competition_type_id"><SelectTrigger><SelectValue placeholder="Тип" /></SelectTrigger><SelectContent>{references.data?.types.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
              <Select name="competition_level_id"><SelectTrigger><SelectValue placeholder="Уровень" /></SelectTrigger><SelectContent>{references.data?.levels.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name ?? item.code ?? `#${item.id}`}</SelectItem>)}</SelectContent></Select>
            </div>
            <Label htmlFor="competition-format">Формат</Label><Select name="format" value={format} onValueChange={setFormat}><SelectTrigger id="competition-format"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="offline">Очно</SelectItem><SelectItem value="online">Онлайн</SelectItem><SelectItem value="mixed">Смешанный</SelectItem></SelectContent></Select>
            <div className="grid gap-3 sm:grid-cols-2"><div><Label htmlFor="start-at">Начало</Label><DateTimePicker id="start-at" name="start_at" required /></div><div><Label htmlFor="end-at">Окончание</Label><DateTimePicker id="end-at" name="end_at" required /></div><div><Label htmlFor="registration-starts-at">Начало регистрации</Label><DateTimePicker id="registration-starts-at" name="registration_starts_at" required /></div><div><Label htmlFor="registration-ends-at">Конец регистрации</Label><DateTimePicker id="registration-ends-at" name="registration_ends_at" required /></div></div>
            <Label htmlFor="venue-address">Адрес площадки{format !== "online" ? " *" : ""}</Label><Input id="venue-address" name="venue_address" required={format !== "online"} placeholder="Адрес площадки…" /><Label htmlFor="online-url">Ссылка на онлайн-площадку{format !== "offline" ? " *" : ""}</Label><Input id="online-url" name="online_url" type="url" required={format !== "offline"} placeholder="https://…" />
            {references.error ? <DataError message={references.error} retry={references.reload} /> : null}
            <DialogFooter><Button type="submit" disabled={pending || references.loading || Boolean(references.error)}>{pending ? <LoaderCircle className="animate-spin" /> : null}Создать</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
