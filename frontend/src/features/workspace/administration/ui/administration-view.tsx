"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import { Activity, Building2, LoaderCircle, Medal, ShieldCheck } from "lucide-react";

import { antiDopingApi, type AntiDopingCertificate } from "@/entities/anti-doping";
import { fileApi } from "@/entities/file";
import { auditLogApi } from "@/entities/audit-log";
import { municipalityApi, type Municipality } from "@/entities/municipality";
import { profileApi } from "@/entities/profile";
import { rankApi, type Rank } from "@/entities/rank";
import { rbacApi, type Permission, type Role } from "@/entities/rbac";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { downloadBlob } from "@/shared/lib/download";
import { useIsMobile } from "@/hooks/use-mobile";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";
import { DatePicker } from "@/shared/ui/date-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Checkbox } from "@/shared/ui/checkbox";
import { Switch } from "@/shared/ui/switch";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Drawer, DrawerContent, DrawerDescription, DrawerFooter, DrawerHandle, DrawerHeader, DrawerTitle } from "@/shared/ui/drawer";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { ButtonGroup } from "@/shared/ui/button-group";

interface AdministrationData {
  roles: Role[];
  permissions: Permission[];
  municipalities: Municipality[];
  ranks: Rank[];
  audit: Awaited<ReturnType<typeof auditLogApi.list>>;
  certificates: AntiDopingCertificate[];
  certificateUsers: Record<number, string>;
}

export type AdministrationSection = "roles" | "directories" | "reviews" | "audit";

function dataSearch<T extends { name: string }>(items: T[] | undefined, query: string) {
  const normalized = query.trim().toLocaleLowerCase("ru");
  return (items ?? []).filter((item) => !normalized || item.name.toLocaleLowerCase("ru").includes(normalized));
}

export function AdministrationView({ section }: { section: AdministrationSection }) {
  const { identity } = useAuth();
  const isMobile = useIsMobile();
  const notify = useNotifications();
  const permissions = new Set(identity?.permissions ?? []);
  const [pending, setPending] = useState<string | null>(null);
  const [selectedPermissionCodes, setSelectedPermissionCodes] = useState<string[]>([]);
  const [selectedMunicipality, setSelectedMunicipality] = useState<Municipality | null>(null);
  const [selectedRank, setSelectedRank] = useState<Rank | null>(null);
  const [roleSearch, setRoleSearch] = useState("");
  const [createRoleOpen, setCreateRoleOpen] = useState(false);
  const [directorySearch, setDirectorySearch] = useState("");
  const [createMunicipalityOpen, setCreateMunicipalityOpen] = useState(false);
  const [createRankOpen, setCreateRankOpen] = useState(false);
  const [directoryMode, setDirectoryMode] = useState<"municipalities" | "ranks">("municipalities");
  const [auditSearch, setAuditSearch] = useState("");

  const load = useCallback(async (): Promise<AdministrationData> => {
    const [roles, permissionItems, municipalities, ranks, audit, certificates] = await Promise.all([
      identity?.permissions?.includes("roles.read") ? rbacApi.roles() : Promise.resolve([]),
      identity?.permissions?.includes("permissions.read") ? rbacApi.permissions() : Promise.resolve([]),
      municipalityApi.list({ limit: 100 }),
      rankApi.list(),
      identity?.permissions?.includes("roles.read") ? auditLogApi.list({ limit: 100 }) : Promise.resolve([]),
      identity?.permissions?.includes("anti_doping.review") ? antiDopingApi.adminList() : Promise.resolve([]),
    ]);
    const userIds = [...new Set(certificates.map((item) => item.user_id))];
    const profiles = await Promise.allSettled(userIds.map((userId) => profileApi.public(userId)));
    const certificateUsers = Object.fromEntries(profiles.flatMap((item, index) => item.status === "fulfilled" ? [[userIds[index], [item.value.last_name, item.value.first_name, item.value.middle_name].filter(Boolean).join(" ")]] : []));
    return { roles, permissions: permissionItems, municipalities, ranks, audit, certificates, certificateUsers };
  }, [identity?.permissions]);
  const state = useAsyncData(load);
  const visibleRoles = useMemo(() => {
    const query = roleSearch.trim().toLocaleLowerCase("ru");
    return (state.data?.roles ?? []).filter((role) => `${role.name} ${role.code} ${role.permissions.join(" ")}`.toLocaleLowerCase("ru").includes(query));
  }, [roleSearch, state.data?.roles]);
  const visibleMunicipalities = useMemo(() => dataSearch(state.data?.municipalities, directorySearch), [directorySearch, state.data?.municipalities]);
  const visibleRanks = useMemo(() => dataSearch(state.data?.ranks, directorySearch), [directorySearch, state.data?.ranks]);
  const visibleAudit = useMemo(() => {
    const query = auditSearch.trim().toLocaleLowerCase("ru");
    return (state.data?.audit ?? []).filter((entry) => `${entry.event_type} ${entry.method} ${entry.route} ${entry.entity_type} ${entry.actor_user_id}`.toLocaleLowerCase("ru").includes(query));
  }, [auditSearch, state.data?.audit]);

  async function run(key: string, action: () => Promise<unknown>, success: string) {
    setPending(key);
    try {
      await action();
      notify.success(success);
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось выполнить действие");
    } finally {
      setPending(null);
    }
  }

  function createRole(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(
      "role",
      () =>
        rbacApi.createRole({
          code: String(form.get("code") ?? "").trim(),
          name: String(form.get("name") ?? "").trim(),
          self_assignable: form.get("self_assignable") === "on",
          permission_codes: selectedPermissionCodes,
        }),
      "Роль создана",
    );
    setCreateRoleOpen(false);
    setSelectedPermissionCodes([]);
  }

  function createMunicipality(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(
      "municipality",
      () => municipalityApi.create({ name: String(form.get("name") ?? "").trim() }),
      "Муниципалитет создан",
    );
    setCreateMunicipalityOpen(false);
  }

  function createRank(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(
      "rank",
      () => rankApi.create({ name: String(form.get("name") ?? "").trim(), ttl: Number(form.get("ttl")) }),
      "Разряд создан",
    );
    setCreateRankOpen(false);
  }

  function reviewCertificate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const reason = String(form.get("rejection_reason") ?? "").trim();
    const completedAt = String(form.get("completed_at") ?? "");
    void run(
      "certificate",
      () =>
        antiDopingApi.review(Number(form.get("certificate_id")), {
          status: String(form.get("status")) as "valid" | "expired" | "rejected",
          completed_at: completedAt || undefined,
          rejection_reason: reason || undefined,
        }),
      "Решение по сертификату сохранено",
    );
  }

  async function downloadCertificate(fileId: number) {
    try {
      const file = await fileApi.download(fileId);
      downloadBlob(file.blob, file.filename);
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось скачать сертификат");
    }
  }

  if (state.loading) return <DataLoading rows={6} />;
  if (state.error) return <DataError message={state.error} retry={state.reload} />;
  const data = state.data;
  if (!data) return null;

  return (
    <div className="mx-auto grid w-full max-w-6xl gap-6">

      {section === "roles" && (permissions.has("roles.read") || permissions.has("roles.create")) ? (
        <div>
          <WorkspaceToolbar search={roleSearch} onSearchChange={setRoleSearch} searchPlaceholder="Найти роль или разрешение" createLabel="Создать роль" onCreate={permissions.has("roles.create") ? () => setCreateRoleOpen(true) : undefined} />
          {!visibleRoles.length ? <Card className="border-0 bg-white/75 p-8 text-center shadow-none"><p className="font-medium">Роли не найдены</p><p className="mt-1 text-sm text-muted-foreground">Измените поисковый запрос.</p></Card> : null}
          <div className="grid gap-3">{visibleRoles.map((role) => <Card key={role.id} className="border-0 bg-white/75 p-4 shadow-none sm:p-5"><div className="flex items-start gap-3"><ShieldCheck className="mt-0.5 size-5 shrink-0 text-primary" /><div className="min-w-0"><p className="font-medium">{role.name}</p><p className="mt-1 text-xs text-muted-foreground">#{role.id} · {role.code}</p></div></div></Card>)}</div>

          {isMobile ? <Drawer open={createRoleOpen} onOpenChange={setCreateRoleOpen}><DrawerContent><DrawerHandle /><DrawerHeader><DrawerTitle>Новая роль</DrawerTitle><DrawerDescription>Задайте название, код и разрешения роли.</DrawerDescription></DrawerHeader><form className="min-h-0 flex-1 grid gap-4 overflow-y-auto scroll-mask-y px-4 pt-4 pb-4" onSubmit={createRole}><Input name="code" placeholder="Код роли" required /><Input name="name" placeholder="Название" required /><label className="flex items-center justify-between gap-4 text-sm"><span>Доступна для самостоятельного выбора</span><Switch name="self_assignable" /></label><div className="grid max-h-48 gap-2 overflow-auto scroll-mask-y rounded-xl bg-muted/50 p-3">{data.permissions.map((permission) => <label key={permission.id} className="flex items-start gap-2 text-sm"><Checkbox checked={selectedPermissionCodes.includes(permission.code)} onCheckedChange={(checked) => setSelectedPermissionCodes((current) => checked ? [...current, permission.code] : current.filter((code) => code !== permission.code))} /><span><b>{permission.code}</b><span className="block text-xs text-muted-foreground">{permission.description}</span></span></label>)}</div><DrawerFooter className="p-0"><Button type="submit" disabled={pending !== null}>{pending === "role" ? <LoaderCircle className="animate-spin" /> : <ShieldCheck />}Создать роль</Button></DrawerFooter></form></DrawerContent></Drawer> : <Dialog open={createRoleOpen} onOpenChange={setCreateRoleOpen}><DialogContent><DialogHeader><DialogTitle>Новая роль</DialogTitle><DialogDescription>Задайте название, код и разрешения роли.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createRole}><Input name="code" placeholder="Код роли" required /><Input name="name" placeholder="Название" required /><label className="flex items-center justify-between gap-4 text-sm"><span>Доступна для самостоятельного выбора</span><Switch name="self_assignable" /></label><div className="grid max-h-48 gap-2 overflow-auto scroll-mask-y rounded-xl bg-muted/50 p-3">{data.permissions.map((permission) => <label key={permission.id} className="flex items-start gap-2 text-sm"><Checkbox checked={selectedPermissionCodes.includes(permission.code)} onCheckedChange={(checked) => setSelectedPermissionCodes((current) => checked ? [...current, permission.code] : current.filter((code) => code !== permission.code))} /><span><b>{permission.code}</b><span className="block text-xs text-muted-foreground">{permission.description}</span></span></label>)}</div><DialogFooter><Button type="submit" disabled={pending !== null}>{pending === "role" ? <LoaderCircle className="animate-spin" /> : <ShieldCheck />}Создать роль</Button></DialogFooter></form></DialogContent></Dialog>}

        </div>
      ) : null}

      {section === "directories" ? <div>
        <ButtonGroup className="mb-3 w-full sm:w-fit [&>[data-slot=button]]:flex-1 sm:[&>[data-slot=button]]:flex-none" aria-label="Раздел справочников"><Button variant="outline" className={directoryMode === "municipalities" ? "bg-primary text-primary-foreground hover:bg-primary/90" : ""} onClick={() => setDirectoryMode("municipalities")}><Building2 />Муниципалитеты</Button><Button variant="outline" className={directoryMode === "ranks" ? "bg-primary text-primary-foreground hover:bg-primary/90" : ""} onClick={() => setDirectoryMode("ranks")}><Medal />Разряды</Button></ButtonGroup>
        <WorkspaceToolbar search={directorySearch} onSearchChange={setDirectorySearch} searchPlaceholder="Найти в справочниках" createLabel={directoryMode === "municipalities" ? "Добавить муниципалитет" : "Добавить разряд"} onCreate={directoryMode === "municipalities" ? (permissions.has("municipalities.create") ? () => setCreateMunicipalityOpen(true) : undefined) : (permissions.has("rank_applications.review") ? () => setCreateRankOpen(true) : undefined)} />
        {directoryMode === "municipalities" ? <Card className="border-0 bg-white/75 p-5 shadow-none"><SectionHeading title="Муниципалитеты" description="Публичный справочник территорий." /><div className="mt-4 grid max-h-80 gap-2 overflow-auto scroll-mask-y">{visibleMunicipalities.map((item) => <button key={item.id} type="button" className="rounded-xl bg-muted/60 p-3 text-left text-sm hover:bg-muted" onClick={() => void municipalityApi.get(item.id).then(setSelectedMunicipality).catch((cause) => notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить муниципалитет"))}>{item.name} <span className="text-muted-foreground">#{item.id}</span></button>)}</div>{!visibleMunicipalities.length ? <p className="mt-4 text-sm text-muted-foreground">Муниципалитеты не найдены.</p> : null}{selectedMunicipality ? <p className="mt-3 text-sm">Выбран: {selectedMunicipality.name} (ID {selectedMunicipality.id})</p> : null}</Card> : <Card className="border-0 bg-white/75 p-5 shadow-none"><SectionHeading title="Спортивные разряды" description="Справочник разрядов и срок действия." /><div className="mt-4 grid max-h-80 gap-2 overflow-auto scroll-mask-y">{visibleRanks.map((item) => <button key={item.id} type="button" className="rounded-xl bg-muted/60 p-3 text-left text-sm hover:bg-muted" onClick={() => void rankApi.get(item.id).then(setSelectedRank).catch((cause) => notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить разряд"))}>{item.name} <span className="text-muted-foreground">#{item.id} · {item.ttl ? `${item.ttl} г.` : "бессрочно"}</span></button>)}</div>{!visibleRanks.length ? <p className="mt-4 text-sm text-muted-foreground">Разряды не найдены.</p> : null}{selectedRank ? <p className="mt-3 text-sm">{selectedRank.name}: создан {formatDate(selectedRank.created_at)}</p> : null}</Card>}
        <Dialog open={createMunicipalityOpen} onOpenChange={setCreateMunicipalityOpen}><DialogContent><DialogHeader><DialogTitle>Новый муниципалитет</DialogTitle><DialogDescription>Добавьте территорию в публичный справочник.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createMunicipality}><Input name="name" placeholder="Название муниципалитета" required /><DialogFooter><Button type="submit" disabled={pending !== null}>{pending === "municipality" ? <LoaderCircle className="animate-spin" /> : <Building2 />}Добавить</Button></DialogFooter></form></DialogContent></Dialog>
        <Dialog open={createRankOpen} onOpenChange={setCreateRankOpen}><DialogContent><DialogHeader><DialogTitle>Новый спортивный разряд</DialogTitle><DialogDescription>Укажите название и срок действия разряда.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createRank}><Input name="name" placeholder="Название" required /><Input name="ttl" type="number" min="0" placeholder="Срок действия, лет" required /><DialogFooter><Button type="submit" disabled={pending !== null}>{pending === "rank" ? <LoaderCircle className="animate-spin" /> : <Medal />}Добавить разряд</Button></DialogFooter></form></DialogContent></Dialog>
      </div> : null}

      {section === "reviews" && permissions.has("anti_doping.review") ? (
        <div className="grid gap-4">
          <SectionHeading title="Проверка антидопинговых сертификатов" description="Очередь сертификатов загружается из административного API." />
          {!data.certificates.length ? <Card className="border-0 bg-white/75 p-8 text-center shadow-none"><p className="font-medium">Сертификатов на проверку нет</p></Card> : null}
          {data.certificates.map((certificate) => <Card key={certificate.id} className="border-0 bg-white/75 p-5 shadow-none"><div className="mb-4 flex flex-wrap items-center justify-between gap-2"><div><p className="font-medium">{data.certificateUsers[certificate.user_id] ?? "Антидопинговый сертификат"}</p><p className="text-sm text-muted-foreground">Статус: {certificate.status}</p></div><Button type="button" size="sm" variant="outline" onClick={() => void downloadCertificate(certificate.file_id)}>Скачать файл</Button></div><form className="grid gap-3 md:grid-cols-2 xl:grid-cols-4" onSubmit={reviewCertificate}><input type="hidden" name="certificate_id" value={certificate.id} /><Select name="status" defaultValue="valid"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="valid">Действителен</SelectItem><SelectItem value="expired">Истёк</SelectItem><SelectItem value="rejected">Отклонён</SelectItem></SelectContent></Select><DatePicker id={`certificate-completed-at-${certificate.id}`} name="completed_at" /><Input name="rejection_reason" placeholder="Причина отклонения" /><Button disabled={pending !== null}>{pending === "certificate" ? <LoaderCircle className="animate-spin" /> : null}Сохранить решение</Button></form></Card>)}
        </div>
      ) : null}

      {section === "audit" && permissions.has("roles.read") ? (
        <div>
          <WorkspaceToolbar search={auditSearch} onSearchChange={setAuditSearch} searchPlaceholder="Найти событие аудита" />
          {!visibleAudit.length ? <Card className="border-0 bg-white/75 p-8 text-center shadow-none"><p className="font-medium">События не найдены</p><p className="mt-1 text-sm text-muted-foreground">Измените поисковый запрос.</p></Card> : null}
          <div className="grid max-h-[32rem] gap-3 overflow-auto scroll-mask-y">{visibleAudit.map((entry) => <Card key={entry.id} className="border-0 bg-white/75 p-4 shadow-none sm:p-5"><div className="flex items-start gap-3"><Activity className="mt-0.5 size-4 shrink-0 text-primary" /><div className="min-w-0 flex-1"><div className="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between"><p className="font-medium">{entry.event_type}</p><p className="text-xs text-muted-foreground">{formatDate(entry.occurred_at, true)}</p></div><p className="mt-1 break-words text-sm text-muted-foreground">{entry.method} {entry.route}</p><p className="mt-2 text-xs text-muted-foreground">Пользователь #{entry.actor_user_id} · {entry.entity_type}{entry.entity_id ? ` #${entry.entity_id}` : ""} · HTTP {entry.status}</p></div></div></Card>)}</div>
        </div>
      ) : null}
    </div>
  );
}
