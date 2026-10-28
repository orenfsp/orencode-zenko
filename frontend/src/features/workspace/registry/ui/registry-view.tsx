"use client";

import { FormEvent, useCallback, useState } from "react";
import { Building2, Search, UserRound } from "lucide-react";

import {
  registryApi,
  type RegistryOrganization,
  type RegistryPage,
  type RegistryUser,
} from "@/entities/registry";
import { formatDate, fullName, SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";

type RegistryMode = "users" | "organizations";
type RegistryData =
  | { mode: "users"; page: RegistryPage<RegistryUser> }
  | { mode: "organizations"; page: RegistryPage<RegistryOrganization> };

export function RegistryView() {
  const [mode, setMode] = useState<RegistryMode>("users");
  const [draftSearch, setDraftSearch] = useState("");
  const [search, setSearch] = useState("");
  const load = useCallback<() => Promise<RegistryData>>(
    async () =>
      mode === "users"
        ? { mode, page: await registryApi.users({ search: search || undefined, limit: 100 }) }
        : { mode, page: await registryApi.organizations({ search: search || undefined, limit: 100 }) },
    [mode, search],
  );
  const state = useAsyncData(load);

  function submitSearch(event: FormEvent) {
    event.preventDefault();
    setSearch(draftSearch.trim());
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading title="Единый реестр" description="Пользователи и организации федерации." />
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-2">
          <Button variant={mode === "users" ? "default" : "outline"} onClick={() => setMode("users")}><UserRound />Пользователи</Button>
          <Button variant={mode === "organizations" ? "default" : "outline"} onClick={() => setMode("organizations")}><Building2 />Организации</Button>
        </div>
        <form onSubmit={submitSearch} className="flex max-w-md gap-2">
          <Input value={draftSearch} onChange={(event) => setDraftSearch(event.target.value)} placeholder="Поиск по реестру" className="bg-white/75 shadow-none" />
          <Button type="submit" variant="outline" size="icon" aria-label="Найти"><Search /></Button>
        </form>
      </div>

      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.page.items.length ? <DataEmpty title="Записей не найдено" description="Попробуйте изменить поисковый запрос." /> : null}

      <div className="grid gap-3">
        {state.data?.mode === "users"
          ? state.data.page.items.map((item) => (
                <Card key={item.id} className="bg-white/75 p-5 shadow-none">
                  <div className="flex flex-col justify-between gap-3 md:flex-row md:items-start">
                    <div>
                      <p className="font-medium">{fullName(item.last_name, item.first_name, item.middle_name)}</p>
                      <p className="mt-1 text-sm text-muted-foreground">{item.organization?.name ?? "Без организации"} · {formatDate(item.birth)}</p>
                    </div>
                    <div className="flex flex-wrap gap-1.5">{item.roles.map((role) => <Badge key={role.id} variant="secondary">{role.name}</Badge>)}</div>
                  </div>
                </Card>
              ))
          : null}
        {state.data?.mode === "organizations"
          ? state.data.page.items.map((item) => (
                <Card key={item.id} className="bg-white/75 p-5 shadow-none">
                  <p className="font-medium">{item.name}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{item.organization_type} · {item.municipality?.name ?? "Муниципалитет не указан"}</p>
                </Card>
              ))
          : null}
      </div>
    </div>
  );
}
