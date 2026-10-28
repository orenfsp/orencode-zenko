"use client";

import { FormEvent, useCallback, useState } from "react";
import { Building2, Download, Search, UserRound } from "lucide-react";

import {
  registryApi,
  type RegistryOrganization,
  type RegistryPage,
  type RegistryUser,
} from "@/entities/registry";
import { formatDate, fullName } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";
import { useNotifications } from "@/shared/lib/notifications";

type RegistryMode = "users" | "organizations";
type RegistryData =
  | { mode: "users"; page: RegistryPage<RegistryUser> }
  | { mode: "organizations"; page: RegistryPage<RegistryOrganization> };

export function RegistryView() {
  const notify = useNotifications();
  const [downloading, setDownloading] = useState(false);
  const [mode, setMode] = useState<RegistryMode>("users");
  const [draftSearch, setDraftSearch] = useState("");
  const [search, setSearch] = useState("");
  const [offset, setOffset] = useState(0);
  const load = useCallback<() => Promise<RegistryData>>(
    async () =>
      mode === "users"
        ? { mode, page: await registryApi.users({ search: search || undefined, limit: 100, offset }) }
        : { mode, page: await registryApi.organizations({ search: search || undefined, limit: 100, offset }) },
    [mode, search, offset],
  );
  const state = useAsyncData(load);

  function submitSearch(event: FormEvent) {
    event.preventDefault();
    setSearch(draftSearch.trim());
    setOffset(0);
  }

  async function downloadReport() {
    setDownloading(true);
    try {
      const report = await registryApi.report();
      const url = URL.createObjectURL(report.blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = report.filename;
      link.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось скачать отчёт");
    } finally {
      setDownloading(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <div className="mb-4 flex justify-end"><Button variant="outline" disabled={downloading} onClick={() => void downloadReport()}><Download />Скачать XLSX</Button></div>
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-2">
          <Button variant={mode === "users" ? "default" : "outline"} onClick={() => { setMode("users"); setOffset(0); }}><UserRound />Пользователи</Button>
          <Button variant={mode === "organizations" ? "default" : "outline"} onClick={() => { setMode("organizations"); setOffset(0); }}><Building2 />Организации</Button>
        </div>
        <form onSubmit={submitSearch} className="flex max-w-md gap-2">
          <Input value={draftSearch} onChange={(event) => setDraftSearch(event.target.value)} placeholder="Поиск по реестру…" aria-label="Поиск по реестру" className="bg-white/75 shadow-none" />
          <Button type="submit" variant="outline" size="icon" aria-label="Найти"><Search /></Button>
        </form>
      </div>

      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.page.items.length ? <DataEmpty title="Записей не найдено" description="Попробуйте изменить поисковый запрос." /> : null}

      <div className="grid gap-3">
        {!state.loading && !state.error && state.data?.mode === "users"
          ? state.data.page.items.map((item) => (
                <Card key={item.id} className="bg-white/75 p-5 shadow-none">
                  <div className="flex flex-col justify-between gap-3 md:flex-row md:items-start">
                    <div>
                      <p className="font-medium">{fullName(item.last_name, item.first_name, item.middle_name)}</p>
                      <p className="mt-1 text-sm text-muted-foreground">{item.organization?.name ?? "Без организации"} · {formatDate(item.birth)}</p>
                    </div>
                    <div className="flex flex-wrap gap-1.5">{item.roles.map((role) => <Badge key={role.id} variant="secondary">{role.name}</Badge>)}</div>
                  </div>
                  <div className="mt-3 grid gap-1 text-xs text-muted-foreground">
                    {item.coach_relations?.length ? <p>Связи с тренерами и спортсменами: {item.coach_relations.map((relation) => `${relation.related_user_name} (${relation.status})`).join(", ")}</p> : null}
                    {item.judge_assignments?.length ? <p>Судейские назначения: {item.judge_assignments.map((assignment) => `${assignment.competition_name} — ${assignment.position_name}`).join(", ")}</p> : null}
                  </div>
                </Card>
              ))
          : null}
        {!state.loading && !state.error && state.data?.mode === "organizations"
          ? state.data.page.items.map((item) => (
                <Card key={item.id} className="bg-white/75 p-5 shadow-none">
                  <p className="font-medium">{item.name}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{item.organization_type} · {item.municipality?.name ?? "Муниципалитет не указан"}</p>
                </Card>
              ))
          : null}
      </div>
      {state.data && !state.loading && !state.error ? <div className="mt-4 flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground"><span>Показаны {state.data.page.items.length ? offset + 1 : 0}–{offset + state.data.page.items.length}{state.data.page.total ? ` из ${state.data.page.total}` : ""}</span><div className="flex gap-2"><Button type="button" variant="outline" size="sm" disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - 100))}>Назад</Button><Button type="button" variant="outline" size="sm" disabled={state.data.page.items.length < 100 || (state.data.page.total > 0 && offset + 100 >= state.data.page.total)} onClick={() => setOffset(offset + 100)}>Далее</Button></div></div> : null}
    </div>
  );
}
