"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import { LoaderCircle } from "lucide-react";

import { deadlineApi } from "@/entities/deadline";
import { useAuth } from "@/entities/session";
import { formatDate, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { DatePicker } from "@/shared/ui/date-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

export function DeadlinesList() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const canManage = identity?.permissions?.includes("deadlines.manage") ?? false;
  const [pendingId, setPendingId] = useState<number | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const load = useCallback(async () => ({
    personal: await deadlineApi.list(),
    reports: canManage ? await deadlineApi.reports() : [],
  }), [canManage]);
  const state = useAsyncData(load);

  async function createReport(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPendingId(0);
    try {
      await deadlineApi.createReport({ name: String(form.get("name") ?? ""), due_at: String(form.get("due_at") ?? ""), type: String(form.get("type")) as "report" | "governor_award", status: "planned" });
      notify.success("Срок добавлен");
      setCreateOpen(false);
      event.currentTarget.reset();
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось добавить срок");
    } finally { setPendingId(null); }
  }

  const personal = useMemo(() => (state.data?.personal ?? []).filter((item) => item.title.toLocaleLowerCase("ru").includes(search.trim().toLocaleLowerCase("ru")) && (statusFilter === "all" || item.status === statusFilter)), [search, state.data, statusFilter]);

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar search={search} onSearchChange={setSearch} searchPlaceholder="Найти срок" filterActive={statusFilter !== "all"} onResetFilters={() => setStatusFilter("all")} filterContent={<Select value={statusFilter} onValueChange={setStatusFilter}><SelectTrigger><SelectValue placeholder="Статус" /></SelectTrigger><SelectContent><SelectItem value="all">Все статусы</SelectItem><SelectItem value="planned">Запланированные</SelectItem><SelectItem value="completed">Выполненные</SelectItem></SelectContent></Select>} createLabel="Добавить срок" onCreate={canManage ? () => setCreateOpen(true) : undefined} />
      <Dialog open={createOpen} onOpenChange={setCreateOpen}><DialogContent><DialogHeader><DialogTitle>Новый отчётный срок</DialogTitle><DialogDescription>Добавьте срок, который будет отображаться федерации и ответственным пользователям.</DialogDescription></DialogHeader><form className="grid gap-3" onSubmit={createReport}>
            <Input name="name" placeholder="Название отчётного срока" required />
            <Select name="type" defaultValue="report"><SelectTrigger aria-label="Тип срока"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="report">Обычный отчёт</SelectItem><SelectItem value="governor_award">Губернаторская премия</SelectItem></SelectContent></Select>
            <DatePicker id="report-due-at" name="due_at" required allowFuture />
            <DialogFooter><Button disabled={pendingId !== null}>{pendingId === 0 ? <LoaderCircle className="animate-spin" /> : null}Добавить</Button></DialogFooter>
          </form></DialogContent></Dialog>
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !personal.length ? (
        <DataEmpty title={search || statusFilter !== "all" ? "Сроки не найдены" : "Активных сроков нет"} description={search || statusFilter !== "all" ? "Измените поисковый запрос или сбросьте фильтры." : "Новые сроки появятся автоматически."} action={search || statusFilter !== "all" ? <Button variant="outline" size="sm" onClick={() => { setSearch(""); setStatusFilter("all"); }}>Сбросить фильтры</Button> : undefined} />
      ) : null}
      <div className="grid gap-3">
        {personal.map((deadline) => (
          <Card key={`${deadline.type}-${deadline.id}`} variant="glass" className="p-5">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <div>
                  <p className="font-medium">{deadline.title}</p>
                  <p className="mt-1 text-sm text-muted-foreground">До {formatDate(deadline.due_at, true)}</p>
                  {deadline.user_name ? <p className="mt-1 text-xs text-muted-foreground">{deadline.user_name}</p> : null}
                </div>
              </div>
              <StatusBadge status={deadline.status} />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
