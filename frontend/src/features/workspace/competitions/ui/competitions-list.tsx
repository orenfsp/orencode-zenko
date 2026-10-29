"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { ArrowRight, CalendarDays, LoaderCircle, MapPin } from "lucide-react";
import Link from "next/link";

import { competitionApi } from "@/entities/competition";
import type { Competition } from "@/entities/competition/model/types";
import { calendarPlanApi } from "@/entities/calendar-plan";
import { referenceDataApi } from "@/entities/reference-data";
import { useAuth } from "@/entities/session";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select";
import { DateTimeRangePicker } from "@/shared/ui/date-time-range-picker";
import { Label } from "@/shared/ui/label";
import { useNotifications } from "@/shared/lib/notifications";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

const competitionFormats: Record<string, string> = {
  offline: "Очно",
  online: "Онлайн",
  mixed: "Смешанный формат",
};

function registrationLabel(startsAt: string, endsAt: string) {
  const now = Date.now();
  const start = new Date(startsAt).getTime();
  const end = new Date(endsAt).getTime();
  if (!Number.isFinite(start) || !Number.isFinite(end))
    return "Срок регистрации уточняется";
  if (now < start) return `Регистрация с ${formatDate(startsAt)}`;
  if (now >= end) return "Регистрация завершена";
  const days = Math.max(1, Math.ceil((end - now) / 86_400_000));
  return `Осталось ${days} дн. регистрации`;
}

export function CompetitionsList() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const canCreate =
    identity?.permissions?.includes("competitions.create") ?? false;
  const canManage =
    identity?.permissions?.includes("competitions.update") ?? false;
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [formatFilter, setFormatFilter] = useState("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [format, setFormat] = useState("offline");
  const [pending, setPending] = useState(false);
  const [recentDrafts, setRecentDrafts] = useState<Competition[]>([]);
  const draftStorageKey = identity
    ? `workspace-competition-drafts-${identity.userId}`
    : null;
  useEffect(() => {
    if (!draftStorageKey) return;
    let active = true;
    let drafts: Competition[] = [];
    try {
      const stored: unknown = JSON.parse(
        sessionStorage.getItem(draftStorageKey) ?? "[]",
      );
      if (Array.isArray(stored)) {
        drafts = stored.filter(
          (item): item is Competition =>
            typeof item === "object" &&
            item !== null &&
            typeof item.id === "number" &&
            item.status === "draft" &&
            typeof item.name === "string",
        );
      }
    } catch {
      /* Invalid or unavailable tab storage. */
    }
    queueMicrotask(() => {
      if (active) setRecentDrafts(drafts);
    });
    return () => {
      active = false;
    };
  }, [draftStorageKey]);
  const loadList = useCallback(() => competitionApi.list({ limit: 100 }), []);
  const loadReferences = useCallback(async () => {
    const [disciplines, types, levels, calendarPlans] = await Promise.all([
      referenceDataApi.disciplines(),
      referenceDataApi.competitionTypes(),
      referenceDataApi.competitionLevels(),
      identity?.permissions?.includes("calendar_plans.read")
        ? calendarPlanApi.list({ limit: 100 })
        : Promise.resolve([]),
    ]);
    return { disciplines, types, levels, calendarPlans };
  }, [identity?.permissions]);
  const state = useAsyncData(loadList);
  const managedState = useAsyncData(
    useCallback(
      () =>
        canManage
          ? competitionApi.adminList({ limit: 100 })
          : Promise.resolve([]),
      [canManage],
    ),
  );
  const references = useAsyncData(loadReferences);
  const competitions = useMemo(() => {
    const query = search.trim().toLocaleLowerCase("ru");
    return (state.data ?? []).filter((competition) => {
      const matchesSearch =
        !query ||
        [
          competition.name,
          competition.discipline.name,
          competition.competition_type.name,
        ]
          .join(" ")
          .toLocaleLowerCase("ru")
          .includes(query);
      return (
        matchesSearch &&
        (statusFilter === "all" || competition.status === statusFilter) &&
        (formatFilter === "all" || competition.format === formatFilter)
      );
    });
  }, [formatFilter, search, state.data, statusFilter]);
  const visibleManaged = [
    ...(managedState.data ?? []),
    ...recentDrafts.filter(
      (draft) =>
        !(managedState.data ?? []).some((item) => item.id === draft.id),
    ),
  ].filter(
    (competition) =>
      !(state.data ?? []).some((item) => item.id === competition.id) &&
      competition.name
        .toLocaleLowerCase("ru")
        .includes(search.trim().toLocaleLowerCase("ru")),
  );

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const nullable = (name: string) =>
      String(form.get(name) ?? "").trim() || null;
    const timestamps = [
      "start_at",
      "end_at",
      "registration_starts_at",
      "registration_ends_at",
    ].map((name) => new Date(String(form.get(name))).getTime());
    if (timestamps.some((value) => !Number.isFinite(value))) {
      notify.error("Укажите даты соревнования и регистрации");
      return;
    }
    if (
      timestamps[0] >= timestamps[1] ||
      timestamps[2] >= timestamps[3] ||
      timestamps[3] > timestamps[0]
    ) {
      notify.error(
        "Проверьте даты: регистрация должна закончиться до начала соревнования, начало — быть раньше окончания",
      );
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
        registration_starts_at: new Date(
          String(form.get("registration_starts_at")),
        ).toISOString(),
        registration_ends_at: new Date(
          String(form.get("registration_ends_at")),
        ).toISOString(),
        venue_address: nullable("venue_address"),
        online_url: nullable("online_url"),
        calendar_plan_id: Number(form.get("calendar_plan_id")) || null,
        calendar_plan_number: nullable("calendar_plan_number"),
        regulation_url: nullable("regulation_url"),
      });
      if (draftStorageKey) {
        const drafts = [
          created,
          ...recentDrafts.filter((draft) => draft.id !== created.id),
        ];
        setRecentDrafts(drafts);
        try {
          sessionStorage.setItem(draftStorageKey, JSON.stringify(drafts));
        } catch {
          /* Storage may be unavailable. */
        }
      }
      notify.success("Черновик соревнования создан");
      setCreateOpen(false);
      state.reload();
    } catch (cause) {
      notify.error(
        cause instanceof Error
          ? cause.message
          : "Не удалось создать соревнование",
      );
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Найти соревнование"
        filterActive={statusFilter !== "all" || formatFilter !== "all"}
        onResetFilters={() => {
          setStatusFilter("all");
          setFormatFilter("all");
        }}
        filterContent={
          <>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger>
                <SelectValue placeholder="Статус" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Все статусы</SelectItem>
                <SelectItem value="published">Опубликовано</SelectItem>
                <SelectItem value="registration_open">
                  Регистрация открыта
                </SelectItem>
                <SelectItem value="completed">Завершено</SelectItem>
              </SelectContent>
            </Select>
            <Select value={formatFilter} onValueChange={setFormatFilter}>
              <SelectTrigger>
                <SelectValue placeholder="Формат" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Все форматы</SelectItem>
                <SelectItem value="offline">Очно</SelectItem>
                <SelectItem value="online">Онлайн</SelectItem>
                <SelectItem value="mixed">Смешанный</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        createLabel="Создать соревнование"
        onCreate={canCreate ? () => setCreateOpen(true) : undefined}
      />
      {visibleManaged.length ? (
        <section className="mb-6" aria-label="Мои соревнования">
          <h2 className="mb-2 text-sm font-medium">Мои соревнования</h2>
          <div className="grid gap-3">
            {visibleManaged.map((draft) => (
              <Card key={draft.id} className="bg-primary/5 p-4 shadow-none">
                <div className="flex flex-wrap items-center gap-2">
                  <StatusBadge status={draft.status} />
                </div>
                <p className="font-medium">{draft.name}</p>
                <p className="text-sm text-muted-foreground">
                  {formatDate(draft.start_at)} — {formatDate(draft.end_at)}
                </p>
                <Button
                  asChild
                  variant="link"
                  size="sm"
                  className="mt-1 h-auto px-0"
                >
                  <Link href={`/workspace/competitions/${draft.id}`}>
                    Открыть и продолжить
                  </Link>
                </Button>
              </Card>
            ))}
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            Здесь доступны черновики и соревнования, которыми вы можете
            управлять.
          </p>
        </section>
      ) : null}
      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? (
        <DataError message={state.error} retry={state.reload} />
      ) : null}
      {!state.loading &&
      !state.error &&
      !competitions.length &&
      !visibleManaged.length ? (
        <DataEmpty
          title={search ? "Ничего не найдено" : "Соревнований пока нет"}
          description={
            search || statusFilter !== "all" || formatFilter !== "all"
              ? "Попробуйте изменить запрос или сбросить фильтры."
              : "Создайте соревнование или дождитесь публикации новых событий федерации."
          }
          action={
            search || statusFilter !== "all" || formatFilter !== "all" ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearch("");
                  setStatusFilter("all");
                  setFormatFilter("all");
                }}
              >
                Сбросить фильтры
              </Button>
            ) : undefined
          }
        />
      ) : null}

      <div className="grid gap-4 lg:grid-cols-2">
        {competitions.map((competition) => (
          <Card
            key={competition.id}
            className="flex h-full flex-col gap-0 border-0 bg-card p-5 shadow-none sm:p-6"
          >
            {/* Статус и формат */}
            <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
              <StatusBadge status={competition.status} />

              <span className="text-xs text-muted-foreground">
                {competitionFormats[competition.format] ?? competition.format}
              </span>
            </div>

            {/* Дисциплина и название */}
            <div className="flex-1">
              <p className="mb-1.5 text-xs text-muted-foreground">
                {competition.discipline.name}
              </p>

              <h3 className="text-lg font-semibold leading-snug tracking-tight">
                {competition.name}
              </h3>
            </div>

            {/* Дата и место */}
            <div className="mt-5 space-y-2.5">
              <div className="flex items-start gap-2.5 text-sm">
                <CalendarDays className="mt-0.5 size-4 shrink-0 text-muted-foreground" />

                <span>
                  {formatDate(competition.start_at)}
                  {" — "}
                  {formatDate(competition.end_at)}
                </span>
              </div>

              <div className="flex items-start gap-2.5 text-sm">
                <MapPin className="mt-0.5 size-4 shrink-0 text-muted-foreground" />

                <span className="line-clamp-2">
                  {competition.format === "online"
                    ? "Онлайн"
                    : competition.venue_address || "Место уточняется"}
                </span>
              </div>
            </div>

            {/* Кнопка и регистрация */}
            <div className="mt-5 flex flex-wrap items-center justify-between gap-3">
              <Button asChild className="w-full rounded-xl sm:w-auto">
                <Link href={`/workspace/competitions/${competition.id}`}>
                  Подробнее
                  <ArrowRight />
                </Link>
              </Button>

              <span className="text-xs text-muted-foreground">
                {registrationLabel(
                  competition.registration_starts_at,
                  competition.registration_ends_at,
                )}
              </span>
            </div>
          </Card>
        ))}
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Новое соревнование</DialogTitle>
            <DialogDescription>
              Заполните основные параметры. Новый объект будет создан как
              черновик.
            </DialogDescription>
          </DialogHeader>
          <form className="grid gap-3" onSubmit={create}>
            <Label htmlFor="competition-name">Название</Label>
            <Input id="competition-name" name="name" required />
            <Label htmlFor="competition-description">Описание</Label>
            <Input id="competition-description" name="description" />
            <div className="grid gap-3 sm:grid-cols-3">
              <Select name="discipline_id" required>
                <SelectTrigger aria-label="Дисциплина">
                  <SelectValue placeholder="Дисциплина" />
                </SelectTrigger>
                <SelectContent>
                  {references.data?.disciplines.map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>
                      {item.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select name="competition_type_id" required>
                <SelectTrigger aria-label="Тип соревнования">
                  <SelectValue placeholder="Тип" />
                </SelectTrigger>
                <SelectContent>
                  {references.data?.types.map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>
                      {item.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select name="competition_level_id" required>
                <SelectTrigger aria-label="Уровень соревнования">
                  <SelectValue placeholder="Уровень" />
                </SelectTrigger>
                <SelectContent>
                  {references.data?.levels.map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>
                      {item.name ?? item.code ?? `#${item.id}`}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-3 sm:grid-cols-[1fr_180px]">
              <div className="grid gap-2">
                <Label htmlFor="calendar-plan-id">Календарный план</Label>
                <Select name="calendar_plan_id">
                  <SelectTrigger id="calendar-plan-id">
                    <SelectValue
                      placeholder={
                        references.data?.calendarPlans.length
                          ? "Выберите план"
                          : "Без календарного плана"
                      }
                    />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">Без календарного плана</SelectItem>
                    {references.data?.calendarPlans.map((plan) => (
                      <SelectItem key={plan.id} value={String(plan.id)}>
                        {plan.name} · {plan.year}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="calendar-plan-number">Номер в плане</Label>
                <Input
                  id="calendar-plan-number"
                  name="calendar_plan_number"
                  placeholder="Например, 12"
                />
              </div>
            </div>
            <Label htmlFor="regulation-url">Ссылка на положение</Label>
            <Input
              id="regulation-url"
              name="regulation_url"
              type="url"
              placeholder="https://…"
            />
            <Label htmlFor="competition-format">Формат</Label>
            <Select name="format" value={format} onValueChange={setFormat}>
              <SelectTrigger id="competition-format">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="offline">Очно</SelectItem>
                <SelectItem value="online">Онлайн</SelectItem>
                <SelectItem value="mixed">Смешанный</SelectItem>
              </SelectContent>
            </Select>
            <div className="grid gap-3 sm:grid-cols-2">
              <div className="grid gap-2">
                <Label htmlFor="registration-period">Период регистрации</Label>
                <DateTimeRangePicker
                  id="registration-period"
                  startName="registration_starts_at"
                  endName="registration_ends_at"
                  defaultStartTime="09:00"
                  defaultEndTime="23:59"
                  required
                  placeholder="Выберите период регистрации"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="competition-period">Период проведения</Label>
                <DateTimeRangePicker
                  id="competition-period"
                  startName="start_at"
                  endName="end_at"
                  required
                  placeholder="Выберите даты проведения"
                />
              </div>
            </div>
            <Label htmlFor="venue-address">
              Адрес площадки{format !== "online" ? " *" : ""}
            </Label>
            <Input
              id="venue-address"
              name="venue_address"
              required={format !== "online"}
              placeholder="Адрес площадки…"
            />
            <Label htmlFor="online-url">
              Ссылка на онлайн-площадку{format !== "offline" ? " *" : ""}
            </Label>
            <Input
              id="online-url"
              name="online_url"
              type="url"
              required={format !== "offline"}
              placeholder="https://…"
            />
            {references.error ? (
              <DataError message={references.error} retry={references.reload} />
            ) : null}
            <DialogFooter>
              <Button
                type="submit"
                disabled={
                  pending || references.loading || Boolean(references.error)
                }
              >
                {pending ? <LoaderCircle className="animate-spin" /> : null}
                Создать
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
