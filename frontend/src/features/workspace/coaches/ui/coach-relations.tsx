"use client";

import { FormEvent, useCallback, useState } from "react";
import { Dumbbell, LoaderCircle, Plus, UserRoundCheck } from "lucide-react";

import { profileApi } from "@/entities/profile";
import type { RegistryUser } from "@/entities/registry";
import { referenceDataApi, type Discipline } from "@/entities/reference-data";
import { useAuth } from "@/entities/session";
import { userCoachApi, type UserCoachRelation } from "@/entities/user-coach";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { ParticipantPicker } from "./participant-picker";

interface RelationData {
  role: "coach" | "participant";
  items: UserCoachRelation[];
  disciplines: Discipline[];
}

export function CoachRelations() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const [pendingId, setPendingId] = useState<number | null>(null);
  const [selectedParticipant, setSelectedParticipant] = useState<RegistryUser | null>(null);
  const [formVersion, setFormVersion] = useState(0);
  const canSearchParticipants = identity?.permissions?.includes("registry.read") ?? false;
  const load = useCallback(async (): Promise<RelationData> => {
    const profile = await profileApi.me();
    const role = profile.roles.find((item) => item.id === identity?.roleId)?.code === "coach" ? "coach" : "participant";
    const [items, disciplines] = await Promise.all([
      role === "coach" ? userCoachApi.byCoach({ limit: 100 }) : userCoachApi.byParticipant({ limit: 100 }),
      referenceDataApi.disciplines(),
    ]);
    return { role, items, disciplines };
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
    <div className="mx-auto w-full max-w-5xl">
      {state.data?.role === "coach" ? (
        <Card className="mb-4 bg-white/75 p-4 shadow-none">
          {canSearchParticipants ? <form key={formVersion} className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]" onSubmit={create}>
            <ParticipantPicker value={selectedParticipant} onChange={setSelectedParticipant} />
            <Select name="discipline_id" required><SelectTrigger className="min-h-11"><SelectValue placeholder="Выберите дисциплину" /></SelectTrigger><SelectContent>{state.data.disciplines.map((discipline) => <SelectItem key={discipline.id} value={String(discipline.id)}>{discipline.name}</SelectItem>)}</SelectContent></Select>
            <Button disabled={pendingId !== null || !selectedParticipant || !state.data.disciplines.length} className="min-h-11">{pendingId === 0 ? <LoaderCircle className="animate-spin" /> : <Plus />}Добавить спортсмена</Button>
          </form> : <p className="text-sm text-muted-foreground">Чтобы найти и привязать спортсмена, нужен доступ к реестру. Обратитесь к администратору федерации.</p>}
        </Card>
      ) : null}
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.items.length ? <DataEmpty title="Связей пока нет" description="Новые запросы появятся на этой странице." /> : null}
      <div className="grid gap-3">
        {state.data?.items.map((relation) => (
          <Card key={relation.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary">
                  {state.data?.role === "coach" ? <Dumbbell className="size-5" /> : <UserRoundCheck className="size-5" />}
                </span>
                <div>
                  <p className="font-medium">{state.data?.role === "coach" ? `Спортсмен #${relation.participant_user_id}` : `Тренер #${relation.coach_user_id}`}</p>
                  <p className="mt-1 text-sm text-muted-foreground">Дисциплина #{relation.discipline_id} · запрос {formatDate(relation.requested_at)}</p>
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
                  <Button variant="outline" size="sm" disabled={pendingId !== null} onClick={() => void update(relation, "end")}>Завершить</Button>
                ) : null}
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
