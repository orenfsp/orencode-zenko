"use client";

import { useCallback, useState, type FormEvent } from "react";
import { Building2, Download, UserRound } from "lucide-react";

import {
  registryApi,
  type RegistryOrganization,
  type RegistryPage,
  type RegistryUser,
} from "@/entities/registry";
import { referenceDataApi } from "@/entities/reference-data";
import { municipalityApi } from "@/entities/municipality";
import { rbacApi } from "@/entities/rbac";
import { useAuth } from "@/entities/session";
import { formatDate, fullName } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { downloadBlob } from "@/shared/lib/download";
import { Badge } from "@/shared/ui/badge";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { useNotifications } from "@/shared/lib/notifications";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { ButtonGroup } from "@/shared/ui/button-group";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Separator } from "@/shared/ui/separator";

type RegistryMode = "users" | "organizations";
type RegistryData =
  | { mode: "users"; page: RegistryPage<RegistryUser> }
  | { mode: "organizations"; page: RegistryPage<RegistryOrganization> };

export function RegistryView() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const [downloading, setDownloading] = useState(false);
  const [mode, setMode] = useState<RegistryMode>("users");
  const [search, setSearch] = useState("");
  const [offset, setOffset] = useState(0);
  const [disciplineId, setDisciplineId] = useState(0);
  const [ageGroupId, setAgeGroupId] = useState(0);
  const [municipalityId, setMunicipalityId] = useState(0);
  const [role, setRole] = useState("all");
  const [selectedUser, setSelectedUser] = useState<RegistryUser | null>(null);
  const [assignOpen, setAssignOpen] = useState(false);
  const [assignPending, setAssignPending] = useState(false);
  const [selectedRoleId, setSelectedRoleId] = useState("");
  const canAssignRole = Boolean(identity?.permissions?.includes("roles.assign") && identity?.permissions?.includes("roles.read"));
  const load = useCallback<() => Promise<RegistryData>>(
    async () =>
      mode === "users"
        ? { mode, page: await registryApi.users({ search: search || undefined, discipline_id: disciplineId || undefined, age_group_id: ageGroupId || undefined, municipality_id: municipalityId || undefined, role: role === "all" ? undefined : role, limit: 100, offset }) }
        : { mode, page: await registryApi.organizations({ search: search || undefined, limit: 100, offset }) },
    [mode, search, offset, disciplineId, ageGroupId, municipalityId, role],
  );
  const state = useAsyncData(load);
  const references = useAsyncData(useCallback(async () => {
    const [disciplines, ageGroups, municipalities, roles] = await Promise.all([
      referenceDataApi.disciplines(),
      referenceDataApi.ageGroups(),
      municipalityApi.list({ limit: 100 }),
      canAssignRole ? rbacApi.roles() : Promise.resolve([]),
    ]);
    return { disciplines, ageGroups, municipalities, roles };
  }, [canAssignRole]));
  const availableRoles = (references.data?.roles ?? []).filter(
    (item) => !selectedUser?.roles.some((role) => role.id === item.id),
  );

  async function assignRole(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedUser || !selectedRoleId) return;
    setAssignPending(true);
    try {
      const roleId = Number(selectedRoleId);
      await rbacApi.assignRole(selectedUser.id, roleId);
      const assignedRole = references.data?.roles.find((item) => item.id === roleId);
      if (assignedRole) {
        setSelectedUser((current) => current ? {
          ...current,
          roles: current.roles.some((item) => item.id === assignedRole.id)
            ? current.roles
            : [...current.roles, assignedRole],
        } : current);
      }
      notify.success("Роль назначена");
      setSelectedRoleId("");
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось назначить роль");
    } finally {
      setAssignPending(false);
    }
  }

  async function downloadReport() {
    setDownloading(true);
    try {
      const report = await registryApi.report();
      downloadBlob(report.blob, report.filename);
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось скачать отчёт");
    } finally {
      setDownloading(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <ButtonGroup className="mb-3 w-full sm:w-fit [&>[data-slot=button]]:flex-1 sm:[&>[data-slot=button]]:flex-none" aria-label="Раздел реестра">
          <Button variant="outline" className={mode === "users" ? "bg-primary text-primary-foreground hover:bg-primary/90" : ""} onClick={() => { setMode("users"); setOffset(0); }}><UserRound />Пользователи</Button>
          <Button variant="outline" className={mode === "organizations" ? "bg-primary text-primary-foreground hover:bg-primary/90" : ""} onClick={() => { setMode("organizations"); setOffset(0); }}><Building2 />Организации</Button>
      </ButtonGroup>
      <div className="fixed inset-x-4 bottom-4 z-30 flex sm:static sm:mb-4"><Button className="w-full sm:w-auto" variant="outline" disabled={downloading} onClick={() => void downloadReport()}><Download />Скачать таблицу пользователей</Button></div>
      <WorkspaceToolbar search={search} onSearchChange={(value) => { setSearch(value); setOffset(0); }} searchPlaceholder="Поиск по реестру…" filterActive={mode === "users" && (disciplineId !== 0 || ageGroupId !== 0 || municipalityId !== 0 || role !== "all")} onResetFilters={() => { setDisciplineId(0); setAgeGroupId(0); setMunicipalityId(0); setRole("all"); setOffset(0); }} filterContent={mode === "users" ? <>
        <Select value={String(disciplineId)} onValueChange={(value) => { setDisciplineId(Number(value)); setOffset(0); }}><SelectTrigger aria-label="Фильтр по дисциплине"><SelectValue placeholder="Все дисциплины" /></SelectTrigger><SelectContent><SelectItem value="0">Все дисциплины</SelectItem>{references.data?.disciplines.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
        <Select value={String(ageGroupId)} onValueChange={(value) => { setAgeGroupId(Number(value)); setOffset(0); }}><SelectTrigger aria-label="Фильтр по возрастной группе"><SelectValue placeholder="Все возрастные группы" /></SelectTrigger><SelectContent><SelectItem value="0">Все возрастные группы</SelectItem>{references.data?.ageGroups.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
        <Select value={String(municipalityId)} onValueChange={(value) => { setMunicipalityId(Number(value)); setOffset(0); }}><SelectTrigger aria-label="Фильтр по муниципалитету"><SelectValue placeholder="Все муниципалитеты" /></SelectTrigger><SelectContent><SelectItem value="0">Все муниципалитеты</SelectItem>{references.data?.municipalities.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
        <Select value={role} onValueChange={(value) => { setRole(value); setOffset(0); }}><SelectTrigger aria-label="Фильтр по роли"><SelectValue placeholder="Все роли" /></SelectTrigger><SelectContent><SelectItem value="all">Все роли</SelectItem><SelectItem value="participant">Спортсмены</SelectItem><SelectItem value="coach">Тренеры</SelectItem><SelectItem value="judge">Судьи</SelectItem></SelectContent></Select>
      </> : undefined} />

      <Dialog open={assignOpen} onOpenChange={(open) => { setAssignOpen(open); if (!open) setSelectedRoleId(""); }}>
        <DialogContent className="max-h-[calc(100svh-2rem)] overflow-y-auto sm:max-w-2xl">
          {selectedUser ? (
            <div className="grid gap-5">
              <DialogHeader>
                <div className="flex items-center gap-4 text-left">
                  <Avatar className="size-16 rounded-2xl">
                    <AvatarFallback className="rounded-[inherit] bg-brand-soft text-primary">
                      <UserRound className="size-7" />
                    </AvatarFallback>
                  </Avatar>
                  <div className="min-w-0">
                    <DialogTitle className="break-words font-mono text-xl leading-tight">
                      {fullName(selectedUser.last_name, selectedUser.first_name, selectedUser.middle_name)}
                    </DialogTitle>
                    <DialogDescription className="mt-1">
                      Профиль пользователя из реестра
                    </DialogDescription>
                  </div>
                </div>
              </DialogHeader>

              <div className="flex flex-wrap gap-2">
                {selectedUser.roles.length ? selectedUser.roles.map((item) => (
                  <Badge key={item.id} variant="secondary">{item.name}</Badge>
                )) : <span className="text-sm text-muted-foreground">Роли не назначены</span>}
              </div>

              <Separator />

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="rounded-xl bg-muted/50 p-4">
                  <p className="text-xs text-muted-foreground">Дата рождения</p>
                  <p className="mt-1 text-sm font-medium">{formatDate(selectedUser.birth)}</p>
                </div>
                <div className="rounded-xl bg-muted/50 p-4">
                  <p className="text-xs text-muted-foreground">Муниципалитет</p>
                  <p className="mt-1 text-sm font-medium">{selectedUser.municipality?.name ?? "Не указан"}</p>
                </div>
                <div className="rounded-xl bg-muted/50 p-4 sm:col-span-2">
                  <p className="text-xs text-muted-foreground">Организация</p>
                  <p className="mt-1 text-sm font-medium">{selectedUser.organization?.name ?? "Не указана"}</p>
                </div>
              </div>

              <section className="grid gap-2">
                <h3 className="text-sm font-medium">Дисциплины</h3>
                <div className="flex flex-wrap gap-2">
                  {selectedUser.disciplines.length ? selectedUser.disciplines.map((item) => (
                    <Badge key={item.id} variant="outline">{item.name}</Badge>
                  )) : <p className="text-sm text-muted-foreground">Дисциплины не указаны</p>}
                </div>
              </section>

              <section className="grid gap-2">
                <h3 className="text-sm font-medium">Спортивные разряды</h3>
                {selectedUser.ranks.length ? (
                  <div className="grid gap-2">
                    {selectedUser.ranks.map((rank) => (
                      <div key={rank.id} className="rounded-xl border p-3">
                        <p className="text-sm font-medium">{rank.rank_name}</p>
                        <p className="mt-1 text-xs text-muted-foreground">{rank.discipline_name}</p>
                      </div>
                    ))}
                  </div>
                ) : <p className="text-sm text-muted-foreground">Подтверждённых разрядов нет</p>}
              </section>

              {selectedUser.judge_category ? (
                <section className="rounded-xl border p-3">
                  <p className="text-xs text-muted-foreground">Судейская категория</p>
                  <p className="mt-1 text-sm font-medium">{selectedUser.judge_category.category_name}</p>
                </section>
              ) : null}

              {canAssignRole ? (
                <form className="grid gap-3 border-t pt-5" onSubmit={assignRole}>
                  <div>
                    <h3 className="text-sm font-medium">Назначить роль</h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Существующие роли сохранятся. Снять роль текущий API не позволяет.
                    </p>
                  </div>
                  <Select value={selectedRoleId} onValueChange={setSelectedRoleId}>
                    <SelectTrigger aria-label="Новая роль пользователя">
                      <SelectValue placeholder={availableRoles.length ? "Выберите роль" : "Все доступные роли уже назначены"} />
                    </SelectTrigger>
                    <SelectContent>
                      {availableRoles.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}
                    </SelectContent>
                  </Select>
                  <DialogFooter>
                    <Button type="submit" disabled={assignPending || !selectedRoleId}>
                      {assignPending ? "Назначение…" : "Назначить роль"}
                    </Button>
                  </DialogFooter>
                </form>
              ) : null}
            </div>
          ) : null}
        </DialogContent>
      </Dialog>

      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.page.items.length ? <DataEmpty title="Записей не найдено" description="Попробуйте изменить поисковый запрос или фильтры." action={search || disciplineId !== 0 || ageGroupId !== 0 || municipalityId !== 0 || role !== "all" ? <Button variant="outline" size="sm" onClick={() => { setSearch(""); setDisciplineId(0); setAgeGroupId(0); setMunicipalityId(0); setRole("all"); setOffset(0); }}>Сбросить фильтры</Button> : undefined} /> : null}

      {!state.loading && !state.error && state.data?.mode === "users" ? (
        <div className="grid gap-3 sm:grid-cols-2">
          {state.data.page.items.map((item) => (
            <Card
              key={item.id}
              role="button"
              tabIndex={0}
              aria-label={`Открыть профиль: ${fullName(item.last_name, item.first_name, item.middle_name)}`}
              className="cursor-pointer border-0 p-5 shadow-none transition-colors hover:bg-accent/80 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              onClick={() => { setSelectedUser(item); setSelectedRoleId(""); setAssignOpen(true); }}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  setSelectedUser(item);
                  setSelectedRoleId("");
                  setAssignOpen(true);
                }
              }}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-medium">{fullName(item.last_name, item.first_name, item.middle_name)}</p>
                  <p className="mt-1 text-xs text-muted-foreground">Дата рождения: {formatDate(item.birth)}</p>
                </div>
                <div className="flex flex-wrap justify-end gap-1.5">
                  {item.roles.map((role) => <Badge key={role.id} variant="secondary">{role.name}</Badge>)}
                </div>
              </div>
              <div className="mt-4 grid gap-1.5 text-sm text-muted-foreground">
                <p>Организация: {item.organization?.name ?? "Не указана"}</p>
                <p>Муниципалитет: {item.municipality?.name ?? "Не указан"}</p>
                <p>Связи: {item.coach_relations?.length ?? 0} · назначения: {item.judge_assignments?.length ?? 0}</p>
              </div>
            </Card>
          ))}
        </div>
      ) : null}
      {!state.loading && !state.error && state.data?.mode === "organizations" ? <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{state.data.page.items.map((item) => <Card key={item.id} className="border-0 p-5 shadow-none"><p className="font-medium">{item.name}</p><div className="mt-4 grid gap-1.5 text-sm text-muted-foreground"><p>Тип: {item.organization_type}</p><p>Муниципалитет: {item.municipality?.name ?? "Не указан"}</p></div></Card>)}</div> : null}
      {state.data && !state.loading && !state.error ? <div className="mt-4 flex flex-wrap items-center justify-between gap-2 pb-16 text-sm text-muted-foreground sm:pb-0"><span>Показаны {state.data.page.items.length ? offset + 1 : 0}–{offset + state.data.page.items.length}{state.data.page.total ? ` из ${state.data.page.total}` : ""}</span><div className="flex gap-2"><Button type="button" variant="outline" size="sm" disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - 100))}>Назад</Button><Button type="button" variant="outline" size="sm" disabled={state.data.page.items.length < 100 || (state.data.page.total > 0 && offset + 100 >= state.data.page.total)} onClick={() => setOffset(offset + 100)}>Далее</Button></div></div> : null}
    </div>
  );
}
