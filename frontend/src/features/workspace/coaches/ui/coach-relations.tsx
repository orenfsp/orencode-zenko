"use client";

import { FormEvent, useCallback, useState } from "react";
import { LoaderCircle, Plus } from "lucide-react";

import { profileApi } from "@/entities/profile";
import { referenceDataApi, type Discipline } from "@/entities/reference-data";
import { useAuth } from "@/entities/session";
import { userCoachApi, type CoachParticipant, type UserCoachRelation } from "@/entities/user-coach";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { ConfirmAction } from "@/shared/ui/confirm-action";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { ParticipantPicker } from "./participant-picker";

interface RelationData {
  role: "coach" | "participant";
  items: UserCoachRelation[];
  disciplines: Discipline[];
  names: Record<number, string>;
}

export function CoachRelations() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const [pendingId, setPendingId] = useState<number | null>(null);
  const [selectedParticipant, setSelectedParticipant] = useState<CoachParticipant | null>(null);
  const [formVersion, setFormVersion] = useState(0);
  const load = useCallback(async (): Promise<RelationData> => {
    const profile = await profileApi.me();
    const role = profile.roles.find((item) => item.id === identity?.roleId)?.code === "coach" ? "coach" : "participant";
    const [items, disciplines] = await Promise.all([
      role === "coach" ? userCoachApi.byCoach({ limit: 100 }) : userCoachApi.byParticipant({ limit: 100 }),
      referenceDataApi.disciplines(),
    ]);
    const userIds = [...new Set(items.map((item) => role === "coach" ? item.participant_user_id : item.coach_user_id))];
    const profiles = await Promise.allSettled(userIds.map((userId) => profileApi.public(userId)));
    const names = Object.fromEntries(profiles.flatMap((profile, index) => profile.status === "fulfilled" ? [[userIds[index], [profile.value.last_name, profile.value.first_name, profile.value.middle_name].filter(Boolean).join(" ")]] : []));
    return { role, items, disciplines, names };
  }, [identity?.roleId]);
  const state = useAsyncData(load);

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedParticipant) return;
    const form = new FormData(event.currentTarget);
    setPendingId(0);
    try {
      await userCoachApi.create({
        participant_user_id: selectedParticipant.id,
        discipline_id: Number(form.get("discipline_id")),
      });
      notify.success("Запрос спортсмену отправлен");
      setSelectedParticipant(null);
      setFormVersion((version) => version + 1);
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось создать связь");
    } finally {
      setPendingId(null);
    }
  }

  async function update(relation: UserCoachRelation, action: "confirm" | "end") {
    setPendingId(relation.id);
    try {
      if (action === "confirm") await userCoachApi.confirm(relation.id);
      else await userCoachApi.end(relation.id);
      notify.success(action === "confirm" ? "Связь подтверждена" : "Связь завершена");
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось изменить связь");
    } finally {
      setPendingId(null);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      {state.data?.role === "coach" ? (
        <Card variant="glass" className="mb-4 p-4">
          <form key={formVersion} className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]" onSubmit={create}>
            <ParticipantPicker value={selectedParticipant} onChange={setSelectedParticipant} />
            <Select name="discipline_id" required><SelectTrigger className="min-h-11"><SelectValue placeholder="Выберите дисциплину" /></SelectTrigger><SelectContent>{state.data.disciplines.map((discipline) => <SelectItem key={discipline.id} value={String(discipline.id)}>{discipline.name}</SelectItem>)}</SelectContent></Select>
            <Button disabled={pendingId !== null || !selectedParticipant || !state.data.disciplines.length} className="min-h-11">{pendingId === 0 ? <LoaderCircle className="animate-spin" /> : <Plus />}Добавить спортсмена</Button>
          </form>
        </Card>
      ) : null}
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.items.length ? <DataEmpty title="Связей пока нет" description="Новые запросы появятся на этой странице." /> : null}
      <div className="grid gap-3">
        {state.data?.items.map((relation) => (
          <Card key={relation.id} variant="glass" className="p-5">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div className="flex gap-3">
                <div>
                  <p className="font-medium">{state.data?.names[state.data.role === "coach" ? relation.participant_user_id : relation.coach_user_id] ?? (state.data?.role === "coach" ? "Спортсмен" : "Тренер")}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{state.data?.disciplines.find((item) => item.id === relation.discipline_id)?.name ?? "Дисциплина не указана"} · запрос {formatDate(relation.requested_at)}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={relation.status} />
                {state.data?.role === "participant" && relation.status === "pending" ? (
                  <Button size="sm" disabled={pendingId !== null} onClick={() => void update(relation, "confirm")}>
                    {pendingId === relation.id ? <LoaderCircle className="animate-spin" /> : null}Подтвердить
                  </Button>
                ) : null}
                {relation.status === "confirmed" ? (
                  <ConfirmAction trigger={<Button variant="outline" size="sm" disabled={pendingId !== null}>Завершить</Button>} title="Завершить связь?" description="Тренер больше не сможет подавать заявки и формировать сборные за этого спортсмена." confirmLabel="Завершить связь" destructive onConfirm={() => update(relation, "end")} />
                ) : null}
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
