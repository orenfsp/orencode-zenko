"use client";

import { FormEvent, useCallback, useState } from "react";
import { LoaderCircle, Pencil, Plus, Search, Trash2, Trophy } from "lucide-react";

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
    <div className="mx-auto w-full max-w-5xl">
      {competitionId ? (
        <div className="mb-4 flex justify-end">
          <Button onClick={() => setEditor({ mode: "create", ageGroupId: null })}>
            <Plus />Создать зачёт
          </Button>
        </div>
      ) : null}

      <Card className="mb-5 bg-white/75 p-5 shadow-none">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
          <div className="flex-1 space-y-2">
            <Label htmlFor="competition-id">Соревнование</Label>
            <Select value={competitionId ? String(competitionId) : ""} onValueChange={(value) => setCompetitionId(Number(value))}>
              <SelectTrigger id="competition-id" className="bg-white shadow-none"><SelectValue placeholder={competitions.loading ? "Загрузка соревнований…" : "Выберите соревнование"} /></SelectTrigger>
              <SelectContent>{competitions.data?.map((competition) => <SelectItem key={competition.id} value={String(competition.id)}>{competition.name} · #{competition.id}</SelectItem>)}</SelectContent>
            </Select>
          </div>
          <Button type="button" variant="outline" disabled={!competitionId} onClick={() => events.reload()}><Search />Открыть</Button>
        </div>
      </Card>

      {events.loading && competitionId ? <DataLoading rows={3} /> : null}
      {events.error ? <DataError message={events.error} retry={events.reload} /> : null}
      {!competitionId ? (
        <DataEmpty title="Выберите соревнование" description="Выберите соревнование из списка, чтобы загрузить его зачёты." />
      ) : null}
      {competitionId && !events.loading && !events.error && !events.data?.length ? (
        <DataEmpty title="Зачётов пока нет" description="Создайте первый зачёт для выбранного соревнования." />
      ) : null}

      <div className="grid gap-3">
        {events.data?.map((item) => (
          <Card key={item.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Trophy className="size-5" /></span>
                <div>
                  <p className="font-medium">Зачёт #{item.id}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{item.discipline.name} · {item.age_group?.name ?? "Без возрастной группы"}</p>
                </div>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" onClick={() => setEditor({ mode: "edit", eventId: item.id, ageGroupId: item.age_group?.id ?? null })}><Pencil />Изменить</Button>
                <Button variant="destructive" size="sm" onClick={() => setDeleting(item)}><Trash2 />Удалить</Button>
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
            <div className="grid max-h-72 gap-2 overflow-y-auto pr-1">
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
