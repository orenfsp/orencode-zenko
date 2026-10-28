"use client";

import { useCallback, useState } from "react";
import { Dumbbell, LoaderCircle, UserRoundCheck } from "lucide-react";

import { profileApi } from "@/entities/profile";
import { useAuth } from "@/entities/session";
import { userCoachApi, type UserCoachRelation } from "@/entities/user-coach";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

interface RelationData {
  role: "coach" | "participant";
  items: UserCoachRelation[];
}

export function CoachRelations() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const [pendingId, setPendingId] = useState<number | null>(null);
  const load = useCallback(async (): Promise<RelationData> => {
    const profile = await profileApi.me();
    const role = profile.roles.find((item) => item.id === identity?.roleId)?.code === "coach" ? "coach" : "participant";
    return { role, items: role === "coach" ? await userCoachApi.byCoach({ limit: 100 }) : await userCoachApi.byParticipant({ limit: 100 }) };
  }, [identity?.roleId]);
  const state = useAsyncData(load);

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
      <SectionHeading
        title={state.data?.role === "coach" ? "Мои спортсмены" : "Мои тренеры"}
        description="Подтверждённые и ожидающие связи по спортивным дисциплинам."
      />
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
