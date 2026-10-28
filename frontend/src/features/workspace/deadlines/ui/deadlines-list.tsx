"use client";

import { FormEvent, useCallback, useState } from "react";
import { CalendarClock, Check, LoaderCircle, Plus } from "lucide-react";

import { deadlineApi } from "@/entities/deadline";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { DatePicker } from "@/shared/ui/date-picker";

export function DeadlinesList() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const canManage = identity?.permissions?.includes("deadlines.manage") ?? false;
  const [pendingId, setPendingId] = useState<number | null>(null);
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
      await deadlineApi.createReport({ name: String(form.get("name") ?? ""), due_at: String(form.get("due_at") ?? ""), status: "planned" });
      notify.success("Срок добавлен");
      event.currentTarget.reset();
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось добавить срок");
    } finally { setPendingId(null); }
  }

  async function completeReport(id: number) {
    setPendingId(id);
    try {
      await deadlineApi.updateReport(id, { status: "completed" });
      notify.success("Срок отмечен выполненным");
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось обновить срок");
    } finally { setPendingId(null); }
  }

  return (
    <div className="mx-auto w-full max-w-5xl">
      <SectionHeading title="Сроки" description="Персональные сроки и обязательные действия для активной роли." />
      {canManage ? (
        <Card className="mb-4 bg-white/75 p-4 shadow-none">
          <form className="grid gap-2 sm:grid-cols-[1fr_180px_auto]" onSubmit={createReport}>
            <Input name="name" placeholder="Название отчётного срока" required />
            <DatePicker id="report-due-at" name="due_at" required />
            <Button disabled={pendingId !== null}>{pendingId === 0 ? <LoaderCircle className="animate-spin" /> : <Plus />}Добавить</Button>
          </form>
        </Card>
      ) : null}
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.personal.length ? (
        <DataEmpty title="Активных сроков нет" description="Новые сроки появятся автоматически." />
      ) : null}
      <div className="grid gap-3">
        {state.data?.personal.map((deadline) => (
          <Card key={`${deadline.type}-${deadline.id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><CalendarClock className="size-5" /></span>
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
      {canManage && state.data?.reports.length ? <>
        <SectionHeading title="Отчётные сроки" description="Сроки, созданные федерацией." />
        <div className="grid gap-3">{state.data.reports.map((report) => <Card key={report.id} className="bg-white/75 p-4 shadow-none"><div className="flex items-center justify-between gap-3"><div><p className="font-medium">{report.name}</p><p className="mt-1 text-sm text-muted-foreground">До {formatDate(report.due_at, true)}</p></div><div className="flex items-center gap-2"><StatusBadge status={report.status} />{report.status === "planned" ? <Button size="sm" variant="outline" disabled={pendingId !== null} onClick={() => void completeReport(report.id)}>{pendingId === report.id ? <LoaderCircle className="animate-spin" /> : <Check />}Выполнено</Button> : null}</div></div></Card>)}</div>
      </> : null}
    </div>
  );
}
