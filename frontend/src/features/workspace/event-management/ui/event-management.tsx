"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import { LoaderCircle, Pencil, Trash2 } from "lucide-react";

import {
  competitionEventApi,
  type CompetitionEvent,
} from "@/entities/competition-event";
import { competitionApi } from "@/entities/competition";
import { referenceDataApi } from "@/entities/reference-data";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Label } from "@/shared/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

interface EventEditor {
  mode: "create" | "edit";
  eventId?: number;
  ageGroupId: number | null;
}

export function EventManagement() {
  const notify = useNotifications();
  const [competitionId, setCompetitionId] = useState<number | null>(null);
  const [editor, setEditor] = useState<EventEditor | null>(null);
  const [deleting, setDeleting] = useState<CompetitionEvent | null>(null);
  const [pending, setPending] = useState(false);
  const [search, setSearch] = useState("");

  const loadEvents = useCallback(
    () =>
      competitionId
        ? competitionEventApi.adminList(competitionId)
        : Promise.resolve([]),
    [competitionId],
  );
  const loadAgeGroups = useCallback(() => referenceDataApi.ageGroups(), []);
  const loadCompetitions = useCallback(() => competitionApi.list({ limit: 100 }), []);
  const events = useAsyncData(loadEvents);
  const ageGroups = useAsyncData(loadAgeGroups);
  const competitions = useAsyncData(loadCompetitions);
  const selectedCompetition = competitions.data?.find((competition) => competition.id === competitionId);
  const visibleEvents = useMemo(() => {
    const query = search.trim().toLocaleLowerCase("ru");
    return (events.data ?? []).filter((item) => `${item.id} ${item.discipline.name} ${item.age_group?.name ?? ""}`.toLocaleLowerCase("ru").includes(query));
  }, [events.data, search]);

  async function saveEvent(event: FormEvent) {
    event.preventDefault();
    if (!competitionId || !editor) return;
    setPending(true);
    try {
      if (editor.mode === "create") {
        await competitionEventApi.create(competitionId, {
          age_group_id: editor.ageGroupId,
        });
        notify.success("Зачёт создан");
      } else if (editor.eventId) {
        await competitionEventApi.update(competitionId, editor.eventId, {
          age_group_id: editor.ageGroupId,
        });
        notify.success("Зачёт обновлён");
      }
      setEditor(null);
      events.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось сохранить зачёт");
    } finally {
      setPending(false);
    }
  }

  async function removeEvent() {
    if (!competitionId || !deleting) return;
    setPending(true);
    try {
      await competitionEventApi.remove(competitionId, deleting.id);
      notify.success("Зачёт удалён");
      setDeleting(null);
      events.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось удалить зачёт");
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Найти зачёт"
        filterContent={
          <div className="grid gap-2">
            <Label htmlFor="competition-id">Соревнование</Label>
            <Select value={competitionId ? String(competitionId) : ""} onValueChange={(value) => setCompetitionId(Number(value))}>
              <SelectTrigger id="competition-id"><SelectValue placeholder={competitions.loading ? "Загрузка соревнований…" : "Выберите соревнование"} /></SelectTrigger>
              <SelectContent>{competitions.data?.map((competition) => <SelectItem key={competition.id} value={String(competition.id)}>{competition.name} · #{competition.id}</SelectItem>)}</SelectContent>
            </Select>
          </div>
        }
        createLabel="Создать зачёт"
        onCreate={competitionId ? () => setEditor({ mode: "create", ageGroupId: null }) : undefined}
      />

      {selectedCompetition ? <p className="mb-4 text-sm text-muted-foreground">Зачёты соревнования: <span className="font-medium text-foreground">{selectedCompetition.name}</span></p> : null}

      {events.loading && competitionId ? <DataLoading rows={3} /> : null}
      {events.error ? <DataError message={events.error} retry={events.reload} /> : null}
      {!competitionId ? (
        <DataEmpty title="Выберите соревнование" description="Выберите соревнование из списка, чтобы загрузить его зачёты." />
      ) : null}
      {competitionId && !events.loading && !events.error && !visibleEvents.length ? (
        <DataEmpty title={search ? "Зачёты не найдены" : "Зачётов пока нет"} description={search ? "Измените поисковый запрос." : "Создайте первый зачёт для выбранного соревнования."} />
      ) : null}

      <div className="grid gap-3">
        {visibleEvents.map((item) => (
          <Card key={item.id} className="border-0 bg-white/75 p-4 shadow-none sm:p-5">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div className="min-w-0">
                <p className="font-medium">Зачёт #{item.id}</p>
                <p className="mt-1 text-sm text-muted-foreground">{item.discipline.name} · {item.age_group?.name ?? "Без возрастной группы"}</p>
              </div>
              <div className="flex gap-2">
                <Button variant="ghost" size="icon" aria-label={`Изменить зачёт №${item.id}`} title="Изменить" onClick={() => setEditor({ mode: "edit", eventId: item.id, ageGroupId: item.age_group?.id ?? null })}><Pencil /></Button>
                <Button variant="ghost" size="icon" className="text-destructive hover:bg-destructive/10 hover:text-destructive" aria-label={`Удалить зачёт №${item.id}`} title="Удалить" onClick={() => setDeleting(item)}><Trash2 /></Button>
              </div>
            </div>
          </Card>
        ))}
      </div>

      <Dialog open={editor !== null} onOpenChange={(open) => { if (!open && !pending) setEditor(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editor?.mode === "create" ? "Новый зачёт" : "Изменение зачёта"}</DialogTitle>
            <DialogDescription>Выберите возрастную группу. Дисциплину backend возьмёт из соревнования.</DialogDescription>
          </DialogHeader>
          <form onSubmit={saveEvent} className="grid gap-4">
            <div className="grid max-h-72 gap-2 overflow-y-auto scroll-mask-y pr-1">
              <Button type="button" variant={editor?.ageGroupId === null ? "default" : "outline"} className="justify-start" onClick={() => setEditor((value) => value ? { ...value, ageGroupId: null } : value)}>Без возрастной группы</Button>
              {ageGroups.data?.map((group) => (
                <Button key={group.id} type="button" variant={editor?.ageGroupId === group.id ? "default" : "outline"} className="justify-start" onClick={() => setEditor((value) => value ? { ...value, ageGroupId: group.id } : value)}>{group.name}</Button>
              ))}
            </div>
            {ageGroups.error ? <p className="text-sm text-destructive">{ageGroups.error}</p> : null}
            <DialogFooter>
              <Button type="button" variant="outline" disabled={pending} onClick={() => setEditor(null)}>Отмена</Button>
              <Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Сохранить</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={deleting !== null} onOpenChange={(open) => { if (!open && !pending) setDeleting(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Удалить зачёт?</DialogTitle>
            <DialogDescription>Операция доступна только для черновика без заявок. Отменить её после удаления нельзя.</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" disabled={pending} onClick={() => setDeleting(null)}>Отмена</Button>
            <Button variant="destructive" disabled={pending} onClick={() => void removeEvent()}>{pending ? <LoaderCircle className="animate-spin" /> : <Trash2 />}Удалить</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
