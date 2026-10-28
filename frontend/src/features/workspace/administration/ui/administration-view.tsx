"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import { Activity, Building2, Check, Download, FileText, LoaderCircle, Medal, ShieldCheck, X } from "lucide-react";

import { antiDopingApi, type AntiDopingCertificate } from "@/entities/anti-doping";
import { fileApi, type UploadedFile } from "@/entities/file";
import { auditLogApi } from "@/entities/audit-log";
import { municipalityApi, type Municipality } from "@/entities/municipality";
import { profileApi } from "@/entities/profile";
import { rankApi, type Rank } from "@/entities/rank";
import { rbacApi, type Permission, type Role } from "@/entities/rbac";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { downloadBlob } from "@/shared/lib/download";
import { useIsMobile } from "@/hooks/use-mobile";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";
import { DatePicker } from "@/shared/ui/date-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Checkbox } from "@/shared/ui/checkbox";
import { Switch } from "@/shared/ui/switch";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Drawer, DrawerContent, DrawerDescription, DrawerFooter, DrawerHandle, DrawerHeader, DrawerTitle } from "@/shared/ui/drawer";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { ButtonGroup } from "@/shared/ui/button-group";
import { Textarea } from "@/shared/ui/textarea";

interface AdministrationData {
  roles: Role[];
  permissions: Permission[];
  municipalities: Municipality[];
  ranks: Rank[];
  audit: Awaited<ReturnType<typeof auditLogApi.list>>;
  certificates: AntiDopingCertificate[];
  certificateUsers: Record<number, string>;
  certificateFiles: Record<number, string>;
}

interface CertificateReviewTarget {
  certificate: AntiDopingCertificate;
  decision: "valid" | "rejected";
}

export type AdministrationSection = "roles" | "directories" | "reviews" | "audit";

function dataSearch<T extends { name: string }>(items: T[] | undefined, query: string) {
  const normalized = query.trim().toLocaleLowerCase("ru");
  return (items ?? []).filter((item) => !normalized || item.name.toLocaleLowerCase("ru").includes(normalized));
}

async function loadAllFiles(): Promise<UploadedFile[]> {
  const result: UploadedFile[] = [];
  const limit = 100;
  for (let offset = 0; ; offset += limit) {
    const page = await fileApi.list({ limit, offset });
    result.push(...page);
    if (page.length < limit) return result;
  }
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
  const [reviewSearch, setReviewSearch] = useState("");
  const [reviewStatus, setReviewStatus] = useState("manual_review");
  const [reviewTarget, setReviewTarget] = useState<CertificateReviewTarget | null>(null);
  const [reviewOpen, setReviewOpen] = useState(false);
  const [reviewLimit, setReviewLimit] = useState(50);

  const load = useCallback(async (): Promise<AdministrationData> => {
    const [roles, permissionItems, municipalities, ranks, audit, certificates, files] = await Promise.all([
      identity?.permissions?.includes("roles.read") ? rbacApi.roles() : Promise.resolve([]),
      identity?.permissions?.includes("permissions.read") ? rbacApi.permissions() : Promise.resolve([]),
      municipalityApi.list({ limit: 100 }),
      rankApi.list(),
      identity?.permissions?.includes("roles.read") ? auditLogApi.list({ limit: 100 }) : Promise.resolve([]),
      section === "reviews" && identity?.permissions?.includes("anti_doping.review") ? antiDopingApi.adminList() : Promise.resolve([]),
      section === "reviews" && identity?.permissions?.includes("files.read") ? loadAllFiles() : Promise.resolve([]),
    ]);
    const userIds = [...new Set(certificates.map((item) => item.user_id))];
    const profiles = await Promise.allSettled(userIds.map((userId) => profileApi.public(userId)));
    const certificateUsers = Object.fromEntries(profiles.flatMap((item, index) => item.status === "fulfilled" ? [[userIds[index], [item.value.last_name, item.value.first_name, item.value.middle_name].filter(Boolean).join(" ")]] : []));
    const certificateFiles = Object.fromEntries(files.map((file) => [file.id, file.original_filename]));
    return { roles, permissions: permissionItems, municipalities, ranks, audit, certificates, certificateUsers, certificateFiles };
  }, [identity?.permissions, section]);
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
  const visibleCertificates = useMemo(() => {
    const query = reviewSearch.trim().toLocaleLowerCase("ru");
    return (state.data?.certificates ?? []).filter((certificate) => {
      const userName = state.data?.certificateUsers[certificate.user_id] ?? "";
      const fileName = state.data?.certificateFiles[certificate.file_id] ?? "";
      const matchesSearch = !query || `${userName} ${fileName}`.toLocaleLowerCase("ru").includes(query);
      return matchesSearch && (reviewStatus === "all" || certificate.status === reviewStatus);
    });
  }, [reviewSearch, reviewStatus, state.data]);

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

  async function reviewCertificate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!reviewTarget) return;
    const form = new FormData(event.currentTarget);
    const reason = String(form.get("rejection_reason") ?? "").trim();
    const completedAt = String(form.get("completed_at") ?? "");
    const pendingKey = `certificate-${reviewTarget.certificate.id}`;
    setPending(pendingKey);
    try {
      await antiDopingApi.review(reviewTarget.certificate.id, {
        status: reviewTarget.decision,
        completed_at: reviewTarget.decision === "valid" ? completedAt : undefined,
        rejection_reason: reviewTarget.decision === "rejected" ? reason : undefined,
      });
      notify.success(reviewTarget.decision === "valid" ? "Сертификат принят" : "Сертификат отклонён");
      setReviewOpen(false);
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось сохранить решение");
    } finally {
      setPending(null);
    }
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

  const reviewForm = reviewTarget ? (
    <form className="grid min-h-0 gap-4 overflow-y-auto px-4 pb-4 md:overflow-visible md:px-0 md:pb-0" onSubmit={reviewCertificate}>
      {reviewTarget.decision === "valid" ? (
        <div className="grid gap-2">
          <label htmlFor="certificate-completed-at" className="text-sm font-medium">Дата прохождения курса</label>
          <DatePicker id="certificate-completed-at" name="completed_at" defaultValue={new Date().toISOString().slice(0, 10)} required />
          <p className="text-xs text-muted-foreground">По правилам backend сертификат будет действовать до 31 декабря указанного года.</p>
        </div>
      ) : (
        <div className="grid gap-2">
          <label htmlFor="certificate-rejection-reason" className="text-sm font-medium">Причина отказа</label>
          <Textarea id="certificate-rejection-reason" name="rejection_reason" maxLength={500} placeholder="Объясните, что нужно исправить или загрузить повторно" required />
          <p className="text-xs text-muted-foreground">Причина будет сохранена в сертификате и доступна пользователю.</p>
        </div>
      )}
      <DialogFooter className="hidden md:flex">
        <Button type="button" variant="outline" onClick={() => setReviewOpen(false)}>Отмена</Button>
        <Button type="submit" variant={reviewTarget.decision === "rejected" ? "destructive" : "default"} disabled={pending !== null}>
          {pending === `certificate-${reviewTarget.certificate.id}` ? <LoaderCircle className="animate-spin" /> : reviewTarget.decision === "valid" ? <Check /> : <X />}
          {reviewTarget.decision === "valid" ? "Принять сертификат" : "Отклонить сертификат"}
        </Button>
      </DialogFooter>
      <DrawerFooter className="p-0 md:hidden">
        <Button type="button" variant="outline" onClick={() => setReviewOpen(false)}>Отмена</Button>
        <Button type="submit" variant={reviewTarget.decision === "rejected" ? "destructive" : "default"} disabled={pending !== null}>
          {pending === `certificate-${reviewTarget.certificate.id}` ? <LoaderCircle className="animate-spin" /> : reviewTarget.decision === "valid" ? <Check /> : <X />}
          {reviewTarget.decision === "valid" ? "Принять сертификат" : "Отклонить сертификат"}
        </Button>
      </DrawerFooter>
    </form>
  ) : null;

  return (
    <div className="mx-auto grid w-full max-w-6xl gap-6">

      {section === "roles" && (permissions.has("roles.read") || permissions.has("roles.create")) ? (
        <div>
          <WorkspaceToolbar search={roleSearch} onSearchChange={setRoleSearch} searchPlaceholder="Найти роль или разрешение" createLabel="Создать роль" onCreate={permissions.has("roles.create") ? () => setCreateRoleOpen(true) : undefined} />
          {!visibleRoles.length ? <Card variant="glass" className="items-center p-8 text-center"><p className="font-medium">Роли не найдены</p><p className="mt-1 text-sm text-muted-foreground">Измените поисковый запрос.</p>{roleSearch ? <Button type="button" variant="outline" size="sm" className="mt-3" onClick={() => setRoleSearch("")}>Сбросить поиск</Button> : null}</Card> : null}
          <div className="grid gap-3">{visibleRoles.map((role) => <Card key={role.id} variant="glass" className="p-4 sm:p-5"><div className="flex items-start gap-3"><ShieldCheck className="mt-0.5 size-5 shrink-0 text-primary" /><div className="min-w-0"><p className="font-medium">{role.name}</p><p className="mt-1 text-xs text-muted-foreground">#{role.id} · {role.code}</p></div></div></Card>)}</div>

          {isMobile ? <Drawer open={createRoleOpen} onOpenChange={setCreateRoleOpen}><DrawerContent><DrawerHandle /><DrawerHeader><DrawerTitle>Новая роль</DrawerTitle><DrawerDescription>Задайте название, код и разрешения роли.</DrawerDescription></DrawerHeader><form className="min-h-0 flex-1 grid gap-4 overflow-y-auto scroll-mask-y px-4 pt-4 pb-4" onSubmit={createRole}><Input name="code" placeholder="Код роли" required /><Input name="name" placeholder="Название" required /><label className="flex items-center justify-between gap-4 text-sm"><span>Доступна для самостоятельного выбора</span><Switch name="self_assignable" /></label><div className="grid max-h-48 gap-2 overflow-auto scroll-mask-y rounded-xl bg-muted/50 p-3">{data.permissions.map((permission) => <label key={permission.id} className="flex items-start gap-2 text-sm"><Checkbox checked={selectedPermissionCodes.includes(permission.code)} onCheckedChange={(checked) => setSelectedPermissionCodes((current) => checked ? [...current, permission.code] : current.filter((code) => code !== permission.code))} /><span><b>{permission.code}</b><span className="block text-xs text-muted-foreground">{permission.description}</span></span></label>)}</div><DrawerFooter className="p-0"><Button type="submit" disabled={pending !== null}>{pending === "role" ? <LoaderCircle className="animate-spin" /> : <ShieldCheck />}Создать роль</Button></DrawerFooter></form></DrawerContent></Drawer> : <Dialog open={createRoleOpen} onOpenChange={setCreateRoleOpen}><DialogContent><DialogHeader><DialogTitle>Новая роль</DialogTitle><DialogDescription>Задайте название, код и разрешения роли.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createRole}><Input name="code" placeholder="Код роли" required /><Input name="name" placeholder="Название" required /><label className="flex items-center justify-between gap-4 text-sm"><span>Доступна для самостоятельного выбора</span><Switch name="self_assignable" /></label><div className="grid max-h-48 gap-2 overflow-auto scroll-mask-y rounded-xl bg-muted/50 p-3">{data.permissions.map((permission) => <label key={permission.id} className="flex items-start gap-2 text-sm"><Checkbox checked={selectedPermissionCodes.includes(permission.code)} onCheckedChange={(checked) => setSelectedPermissionCodes((current) => checked ? [...current, permission.code] : current.filter((code) => code !== permission.code))} /><span><b>{permission.code}</b><span className="block text-xs text-muted-foreground">{permission.description}</span></span></label>)}</div><DialogFooter><Button type="submit" disabled={pending !== null}>{pending === "role" ? <LoaderCircle className="animate-spin" /> : <ShieldCheck />}Создать роль</Button></DialogFooter></form></DialogContent></Dialog>}

        </div>
      ) : null}

      {section === "directories" ? <div>
        <ButtonGroup className="mb-3 w-full sm:w-fit [&>[data-slot=button]]:flex-1 sm:[&>[data-slot=button]]:flex-none" aria-label="Раздел справочников"><Button variant="outline" className={directoryMode === "municipalities" ? "bg-primary text-primary-foreground hover:bg-primary/90" : ""} onClick={() => setDirectoryMode("municipalities")}><Building2 />Муниципалитеты</Button><Button variant="outline" className={directoryMode === "ranks" ? "bg-primary text-primary-foreground hover:bg-primary/90" : ""} onClick={() => setDirectoryMode("ranks")}><Medal />Разряды</Button></ButtonGroup>
        <WorkspaceToolbar search={directorySearch} onSearchChange={setDirectorySearch} searchPlaceholder="Найти в справочниках" createLabel={directoryMode === "municipalities" ? "Добавить муниципалитет" : "Добавить разряд"} onCreate={directoryMode === "municipalities" ? (permissions.has("municipalities.create") ? () => setCreateMunicipalityOpen(true) : undefined) : (permissions.has("rank_applications.review") ? () => setCreateRankOpen(true) : undefined)} />
        {directoryMode === "municipalities" ? <Card variant="glass" className="p-5"><SectionHeading title="Муниципалитеты" description="Публичный справочник территорий." /><div className="mt-4 grid max-h-80 gap-2 overflow-auto scroll-mask-y">{visibleMunicipalities.map((item) => <Button key={item.id} type="button" variant="ghost" className="h-auto justify-start bg-muted/60 p-3 text-left font-normal hover:bg-muted" onClick={() => void municipalityApi.get(item.id).then(setSelectedMunicipality).catch((cause) => notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить муниципалитет"))}>{item.name} <span className="text-muted-foreground">#{item.id}</span></Button>)}</div>{!visibleMunicipalities.length ? <p className="mt-4 text-sm text-muted-foreground">Муниципалитеты не найдены.</p> : null}{selectedMunicipality ? <p className="mt-3 text-sm">Выбран: {selectedMunicipality.name} (ID {selectedMunicipality.id})</p> : null}</Card> : <Card variant="glass" className="p-5"><SectionHeading title="Спортивные разряды" description="Справочник разрядов и срок действия." /><div className="mt-4 grid max-h-80 gap-2 overflow-auto scroll-mask-y">{visibleRanks.map((item) => <Button key={item.id} type="button" variant="ghost" className="h-auto justify-start bg-muted/60 p-3 text-left font-normal hover:bg-muted" onClick={() => void rankApi.get(item.id).then(setSelectedRank).catch((cause) => notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить разряд"))}>{item.name} <span className="text-muted-foreground">#{item.id} · {item.ttl ? `${item.ttl} г.` : "бессрочно"}</span></Button>)}</div>{!visibleRanks.length ? <p className="mt-4 text-sm text-muted-foreground">Разряды не найдены.</p> : null}{selectedRank ? <p className="mt-3 text-sm">{selectedRank.name}: создан {formatDate(selectedRank.created_at)}</p> : null}</Card>}
        <Dialog open={createMunicipalityOpen} onOpenChange={setCreateMunicipalityOpen}><DialogContent><DialogHeader><DialogTitle>Новый муниципалитет</DialogTitle><DialogDescription>Добавьте территорию в публичный справочник.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createMunicipality}><Input name="name" placeholder="Название муниципалитета" required /><DialogFooter><Button type="submit" disabled={pending !== null}>{pending === "municipality" ? <LoaderCircle className="animate-spin" /> : <Building2 />}Добавить</Button></DialogFooter></form></DialogContent></Dialog>
        <Dialog open={createRankOpen} onOpenChange={setCreateRankOpen}><DialogContent><DialogHeader><DialogTitle>Новый спортивный разряд</DialogTitle><DialogDescription>Укажите название и срок действия разряда.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={createRank}><Input name="name" placeholder="Название" required /><Input name="ttl" type="number" min="0" placeholder="Срок действия, лет" required /><DialogFooter><Button type="submit" disabled={pending !== null}>{pending === "rank" ? <LoaderCircle className="animate-spin" /> : <Medal />}Добавить разряд</Button></DialogFooter></form></DialogContent></Dialog>
      </div> : null}

      {section === "reviews" && permissions.has("anti_doping.review") ? (
        <div>
          <WorkspaceToolbar
            search={reviewSearch}
            onSearchChange={(value) => { setReviewSearch(value); setReviewLimit(50); }}
            searchPlaceholder="Найти пользователя или файл"
            filterActive={reviewStatus !== "manual_review"}
            onResetFilters={() => { setReviewStatus("manual_review"); setReviewLimit(50); }}
            filterContent={(
              <Select value={reviewStatus} onValueChange={(value) => { setReviewStatus(value); setReviewLimit(50); }}>
                <SelectTrigger><SelectValue placeholder="Статус сертификата" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="manual_review">Ожидают проверки</SelectItem>
                  <SelectItem value="valid">Принятые</SelectItem>
                  <SelectItem value="rejected">Отклонённые</SelectItem>
                  <SelectItem value="expired">Истёкшие</SelectItem>
                  <SelectItem value="all">Все сертификаты</SelectItem>
                </SelectContent>
              </Select>
            )}
          />

          {!visibleCertificates.length ? (
            <DataEmpty
              title={reviewSearch || reviewStatus !== "manual_review" ? "Сертификаты не найдены" : "Очередь проверки пуста"}
              description={reviewSearch || reviewStatus !== "manual_review" ? "Измените запрос или сбросьте фильтры." : "Новые сертификаты появятся здесь после загрузки пользователями."}
              action={reviewSearch || reviewStatus !== "manual_review" ? <Button type="button" variant="outline" size="sm" onClick={() => { setReviewSearch(""); setReviewStatus("manual_review"); setReviewLimit(50); }}>Сбросить фильтры</Button> : undefined}
            />
          ) : null}

          <div className="grid gap-3">
            {visibleCertificates.slice(0, reviewLimit).map((certificate) => {
              const userName = data.certificateUsers[certificate.user_id] ?? `Пользователь #${certificate.user_id}`;
              const fileName = data.certificateFiles[certificate.file_id] ?? `Сертификат #${certificate.file_id}`;
              return (
                <Card key={certificate.id} variant="glass" className="p-4 sm:p-5">
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex min-w-0 items-start gap-3">
                      <FileText className="mt-0.5 size-5 shrink-0 text-primary" />
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2"><p className="font-medium">{userName}</p><StatusBadge status={certificate.status} /></div>
                        <p className="mt-1 truncate text-sm text-foreground" title={fileName}>{fileName}</p>
                        <p className="mt-1 text-xs text-muted-foreground">Загружен {formatDate(certificate.created_at, true)}{certificate.valid_until ? ` · действует до ${formatDate(certificate.valid_until)}` : ""}</p>
                      </div>
                    </div>
                    <div className="flex w-full flex-wrap gap-2 sm:w-auto sm:justify-end">
                      <Button type="button" variant="outline" size="sm" className="flex-1 sm:flex-none" onClick={() => void downloadCertificate(certificate.file_id)}><Download />Скачать</Button>
                      {certificate.status === "manual_review" ? <>
                        <Button type="button" variant="outline" size="sm" className="flex-1 sm:flex-none" onClick={() => { setReviewTarget({ certificate, decision: "rejected" }); setReviewOpen(true); }}><X />Отклонить</Button>
                        <Button type="button" size="sm" className="flex-1 sm:flex-none" onClick={() => { setReviewTarget({ certificate, decision: "valid" }); setReviewOpen(true); }}><Check />Принять</Button>
                      </> : null}
                    </div>
                  </div>
                </Card>
              );
            })}
          </div>
          {visibleCertificates.length > reviewLimit ? <div className="mt-4 flex flex-col items-center gap-2"><p className="text-xs text-muted-foreground">Показано {reviewLimit} из {visibleCertificates.length}</p><Button type="button" variant="outline" onClick={() => setReviewLimit((current) => current + 50)}>Показать ещё</Button></div> : null}

          {isMobile ? (
            <Drawer open={reviewOpen} onOpenChange={setReviewOpen}>
              <DrawerContent>
                <DrawerHandle />
                <DrawerHeader>
                  <DrawerTitle>{reviewTarget?.decision === "valid" ? "Принять сертификат" : "Отклонить сертификат"}</DrawerTitle>
                  <DrawerDescription>{reviewTarget ? `${data.certificateUsers[reviewTarget.certificate.user_id] ?? `Пользователь #${reviewTarget.certificate.user_id}`} · ${data.certificateFiles[reviewTarget.certificate.file_id] ?? `Сертификат #${reviewTarget.certificate.file_id}`}` : "Проверьте решение"}</DrawerDescription>
                </DrawerHeader>
                {reviewForm}
              </DrawerContent>
            </Drawer>
          ) : (
            <Dialog open={reviewOpen} onOpenChange={setReviewOpen}>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>{reviewTarget?.decision === "valid" ? "Принять сертификат" : "Отклонить сертификат"}</DialogTitle>
                  <DialogDescription>{reviewTarget ? `${data.certificateUsers[reviewTarget.certificate.user_id] ?? `Пользователь #${reviewTarget.certificate.user_id}`} · ${data.certificateFiles[reviewTarget.certificate.file_id] ?? `Сертификат #${reviewTarget.certificate.file_id}`}` : "Проверьте решение"}</DialogDescription>
                </DialogHeader>
                {reviewForm}
              </DialogContent>
            </Dialog>
          )}
        </div>
      ) : null}

      {section === "audit" && permissions.has("roles.read") ? (
        <div>
          <WorkspaceToolbar search={auditSearch} onSearchChange={setAuditSearch} searchPlaceholder="Найти событие аудита" />
          {!visibleAudit.length ? <Card variant="glass" className="items-center p-8 text-center"><p className="font-medium">События не найдены</p><p className="mt-1 text-sm text-muted-foreground">Измените поисковый запрос.</p>{auditSearch ? <Button type="button" variant="outline" size="sm" className="mt-3" onClick={() => setAuditSearch("")}>Сбросить поиск</Button> : null}</Card> : null}
          <div className="grid max-h-[32rem] gap-3 overflow-auto scroll-mask-y">{visibleAudit.map((entry) => <Card key={entry.id} variant="glass" className="p-4 sm:p-5"><div className="flex items-start gap-3"><Activity className="mt-0.5 size-4 shrink-0 text-primary" /><div className="min-w-0 flex-1"><div className="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between"><p className="font-medium">{entry.event_type}</p><p className="text-xs text-muted-foreground">{formatDate(entry.occurred_at, true)}</p></div><p className="mt-1 break-words text-sm text-muted-foreground">{entry.method} {entry.route}</p><p className="mt-2 text-xs text-muted-foreground">Пользователь #{entry.actor_user_id} · {entry.entity_type}{entry.entity_id ? ` #${entry.entity_id}` : ""} · HTTP {entry.status}</p></div></div></Card>)}</div>
        </div>
      ) : null}
    </div>
  );
}
