"use client";

import { useCallback, useMemo, useState } from "react";
import { Check, LoaderCircle } from "lucide-react";

import { notificationApi } from "@/entities/notification";
import { formatDate } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";

export function NotificationsView() {
  const notify = useNotifications();
  const [pendingId, setPendingId] = useState<number | null>(null);
  const load = useCallback(() => notificationApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);
  const [search, setSearch] = useState("");
  const [readFilter, setReadFilter] = useState("all");
  const items = useMemo(() => (state.data ?? []).filter((item) => `${item.title} ${item.message}`.toLocaleLowerCase("ru").includes(search.trim().toLocaleLowerCase("ru")) && (readFilter === "all" || (readFilter === "unread" ? !item.is_read : item.is_read))), [readFilter, search, state.data]);

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
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar search={search} onSearchChange={setSearch} searchPlaceholder="Найти уведомление" filterActive={readFilter !== "all"} onResetFilters={() => setReadFilter("all")} filterContent={<Select value={readFilter} onValueChange={setReadFilter}><SelectTrigger><SelectValue placeholder="Статус" /></SelectTrigger><SelectContent><SelectItem value="all">Все уведомления</SelectItem><SelectItem value="unread">Непрочитанные</SelectItem><SelectItem value="read">Прочитанные</SelectItem></SelectContent></Select>} />
      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !items.length ? <DataEmpty title={search || readFilter !== "all" ? "Уведомления не найдены" : "Уведомлений пока нет"} action={search || readFilter !== "all" ? <Button variant="outline" size="sm" onClick={() => { setSearch(""); setReadFilter("all"); }}>Сбросить фильтры</Button> : undefined} /> : null}
      <div className="grid gap-3">
        {items.map((item) => {
          const isGenericTitle = item.title.trim().toLocaleLowerCase("ru") === "уведомление";
          const headingClass = item.is_read ? "font-medium text-foreground/55" : "font-medium text-foreground";

          return (
          <Card key={item.id} variant="glass" className="p-4 sm:p-5">
            <div className="relative">
              <div className="min-w-0 pr-8">
                {!isGenericTitle ? <p className={headingClass}>{item.title}</p> : null}
                <p className={isGenericTitle ? `${headingClass} leading-5` : "mt-1 text-sm leading-5 text-muted-foreground"}>{item.message}</p>
                <p className="mt-2 text-xs text-muted-foreground">{formatDate(item.created_at, true)}</p>
              </div>
              {!item.is_read ? (
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label="Отметить уведомление как прочитанное"
                  title="Отметить как прочитанное"
                  disabled={pendingId !== null}
                  onClick={() => void markRead(item.id)}
                  className="absolute right-0 bottom-0 size-8 shrink-0 text-primary hover:bg-primary/10 hover:text-primary"
                >
                  {pendingId === item.id ? <LoaderCircle className="animate-spin" /> : <Check />}
                </Button>
              ) : <Check className="absolute right-1 bottom-1 size-4 text-primary" aria-label="Прочитано" />}
            </div>
          </Card>
          );
        })}
      </div>
    </div>
  );
}
