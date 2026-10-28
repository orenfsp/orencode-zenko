"use client";

import { useCallback, useState } from "react";
import { Bell, Check, LoaderCircle } from "lucide-react";

import { notificationApi } from "@/entities/notification";
import { formatDate, SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function NotificationsView() {
  const notify = useNotifications();
  const [pendingId, setPendingId] = useState<number | null>(null);
  const load = useCallback(() => notificationApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);

  async function markRead(id: number) {
    setPendingId(id);
    try {
      await notificationApi.markRead(id);
      notify.success("Уведомление прочитано");
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось отметить уведомление");
    } finally {
      setPendingId(null);
    }
  }

  return (
    <div className="mx-auto w-full max-w-5xl">
      <SectionHeading title="Уведомления" description="Решения по заявкам и важные события аккаунта." />
      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? <DataEmpty title="Уведомлений пока нет" /> : null}
      <div className="grid gap-3">
        {state.data?.map((item) => (
          <Card key={item.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Bell className="size-5" /></span>
                <div>
                  <p className="font-medium">{item.title}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{item.message}</p>
                  <p className="mt-2 text-xs text-muted-foreground">{formatDate(item.created_at, true)}</p>
                </div>
              </div>
              {!item.is_read ? (
                <Button variant="outline" size="sm" disabled={pendingId !== null} onClick={() => void markRead(item.id)}>
                  {pendingId === item.id ? <LoaderCircle className="animate-spin" /> : <Check />}Прочитано
                </Button>
              ) : <span className="text-xs text-muted-foreground">Прочитано</span>}
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
