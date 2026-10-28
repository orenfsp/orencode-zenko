"use client";

import { FormEvent, useCallback, useDeferredValue, useState } from "react";
import { Activity, Building2, LoaderCircle, Medal, ShieldCheck } from "lucide-react";

import { antiDopingApi } from "@/entities/anti-doping";
import { auditLogApi } from "@/entities/audit-log";
import { municipalityApi, type Municipality } from "@/entities/municipality";
import { registryApi, type RegistryUser } from "@/entities/registry";
import { rankApi, type Rank } from "@/entities/rank";
import { rbacApi, type Permission, type Role } from "@/entities/rbac";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";
import { DatePicker } from "@/shared/ui/date-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Checkbox } from "@/shared/ui/checkbox";

interface AdministrationData {
  roles: Role[];
  permissions: Permission[];
  municipalities: Municipality[];
  ranks: Rank[];
  audit: Awaited<ReturnType<typeof auditLogApi.list>>;
  users: RegistryUser[];
}

export type AdministrationSection = "roles" | "directories" | "reviews" | "audit";

export function AdministrationView({ section }: { section: AdministrationSection }) {
  const { identity } = useAuth();
  const notify = useNotifications();
  const permissions = new Set(identity?.permissions ?? []);
  const [pending, setPending] = useState<string | null>(null);
  const [selectedPermissionCodes, setSelectedPermissionCodes] = useState<string[]>([]);
  const [selectedMunicipality, setSelectedMunicipality] = useState<Municipality | null>(null);
  const [selectedRank, setSelectedRank] = useState<Rank | null>(null);
  const [userSearch, setUserSearch] = useState("");
  const deferredUserSearch = useDeferredValue(userSearch);

  const load = useCallback(async (): Promise<AdministrationData> => {
    const [roles, permissionItems, municipalities, ranks, audit, users] = await Promise.all([
      identity?.permissions?.includes("roles.read") ? rbacApi.roles() : Promise.resolve([]),
      identity?.permissions?.includes("permissions.read") ? rbacApi.permissions() : Promise.resolve([]),
      municipalityApi.list({ limit: 100 }),
      rankApi.list(),
      identity?.permissions?.includes("roles.read") ? auditLogApi.list({ limit: 100 }) : Promise.resolve([]),
      identity?.permissions?.includes("roles.assign") && identity?.permissions?.includes("registry.read")
        ? registryApi.users({ search: deferredUserSearch || undefined, limit: 100 }).then((page) => page.items)
        : Promise.resolve([]),
    ]);
    return { roles, permissions: permissionItems, municipalities, ranks, audit, users };
  }, [identity?.permissions, deferredUserSearch]);
  const state = useAsyncData(load);

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
  }

  function assignRole(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(
      "assign",
      () => rbacApi.assignRole(Number(form.get("user_id")), Number(form.get("role_id"))),
      "Роль назначена",
    );
  }

  function createMunicipality(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(
      "municipality",
      () => municipalityApi.create({ name: String(form.get("name") ?? "").trim() }),
      "Муниципалитет создан",
    );
  }

  function createRank(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    void run(
      "rank",
      () => rankApi.create({ name: String(form.get("name") ?? "").trim(), ttl: Number(form.get("ttl")) }),
      "Разряд создан",
    );
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

  if (state.loading) return <DataLoading rows={6} />;
  if (state.error) return <DataError message={state.error} retry={state.reload} />;
  const data = state.data;
  if (!data) return null;

  return (
    <div className="mx-auto grid w-full max-w-7xl gap-6">
      <SectionHeading title="Администрирование" description="Роли, справочники, проверки и журнал действий." />

      {section === "roles" && (permissions.has("roles.read") || permissions.has("roles.create") || permissions.has("roles.assign")) ? (
        <Card className="bg-white/75 p-5 shadow-none">
          <SectionHeading title="Роли и права" description="Создание ролей и назначение пользователям." />
          <div className="grid gap-5 xl:grid-cols-2">
            <div className="grid content-start gap-2">
              {data.roles.map((role) => (
                <div key={role.id} className="rounded-xl bg-muted/60 p-3">
                  <p className="font-medium">{role.name} <span className="text-xs text-muted-foreground">#{role.id} · {role.code}</span></p>
                  <p className="mt-1 text-xs text-muted-foreground">{role.permissions.join(", ") || "Без разрешений"}</p>
                </div>
              ))}
            </div>
            <div className="grid gap-5">
              {permissions.has("roles.create") ? (
                <form className="grid gap-3 rounded-xl border p-4" onSubmit={createRole}>
                  <p className="font-medium">Новая роль</p>
                  <Input name="code" placeholder="Код роли" required />
                  <Input name="name" placeholder="Название" required />
                  <label className="flex items-center gap-2 text-sm"><Checkbox name="self_assignable" /> Доступна для самостоятельного выбора</label>
                  <div className="grid max-h-40 gap-2 overflow-auto rounded-xl bg-muted/50 p-3">
                    {data.permissions.map((permission) => (
                      <label key={permission.id} className="flex items-start gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={selectedPermissionCodes.includes(permission.code)}
                          onChange={(event) => setSelectedPermissionCodes((current) => event.target.checked ? [...current, permission.code] : current.filter((code) => code !== permission.code))}
                        />
                        <span><b>{permission.code}</b><span className="block text-xs text-muted-foreground">{permission.description}</span></span>
                      </label>
                    ))}
                  </div>
                  <Button disabled={pending !== null}>{pending === "role" ? <LoaderCircle className="animate-spin" /> : <ShieldCheck />}Создать роль</Button>
                </form>
              ) : null}
              {permissions.has("roles.assign") ? (
                <div className="grid gap-3 rounded-xl border p-4">
                  <p className="font-medium">Назначить роль</p>
                  {permissions.has("registry.read") ? <Input value={userSearch} onChange={(event) => setUserSearch(event.target.value)} placeholder="Поиск по имени пользователя" aria-label="Поиск пользователя для назначения роли" /> : <p className="text-sm text-muted-foreground">Для выбора пользователя требуется право просмотра реестра. Обратитесь к администратору.</p>}
                  {!permissions.has("roles.read") ? <p className="text-sm text-muted-foreground">Для выбора роли требуется право просмотра ролей. Обратитесь к администратору.</p> : null}
                  <form className="grid gap-3" onSubmit={assignRole}>
                    {data.users.length ? <Select key={deferredUserSearch} name="user_id" required><SelectTrigger><SelectValue placeholder="Выберите пользователя" /></SelectTrigger><SelectContent>{data.users.map((user) => <SelectItem key={user.id} value={String(user.id)}>{user.last_name} {user.first_name}{user.middle_name ? ` ${user.middle_name}` : ""} · #{user.id}</SelectItem>)}</SelectContent></Select> : permissions.has("registry.read") ? <p className="text-sm text-muted-foreground">Пользователи не найдены. Попробуйте другой запрос.</p> : null}
                    {data.roles.length ? <Select name="role_id" required><SelectTrigger><SelectValue placeholder="Выберите роль" /></SelectTrigger><SelectContent>{data.roles.map((role) => <SelectItem key={role.id} value={String(role.id)}>{role.name}</SelectItem>)}</SelectContent></Select> : null}
                    <Button disabled={pending !== null || !data.users.length || !data.roles.length}>{pending === "assign" ? <LoaderCircle className="animate-spin" /> : null}Назначить</Button>
                  </form>
                </div>
              ) : null}
            </div>
          </div>
        </Card>
      ) : null}

      {section === "directories" ? <div className="grid gap-6 xl:grid-cols-2">
        <Card className="bg-white/75 p-5 shadow-none">
          <SectionHeading title="Муниципалитеты" description="Публичный справочник территорий." />
          {permissions.has("municipalities.create") ? (
            <form className="mb-4 flex gap-2" onSubmit={createMunicipality}><Input name="name" placeholder="Название муниципалитета" required /><Button disabled={pending !== null}>{pending === "municipality" ? <LoaderCircle className="animate-spin" /> : <Building2 />}Добавить</Button></form>
          ) : null}
          <div className="grid max-h-80 gap-2 overflow-auto">
            {data.municipalities.map((item) => (
              <button key={item.id} type="button" className="rounded-xl bg-muted/60 p-3 text-left text-sm hover:bg-muted" onClick={() => void municipalityApi.get(item.id).then(setSelectedMunicipality).catch((cause) => notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить муниципалитет"))}>
                {item.name} <span className="text-muted-foreground">#{item.id}</span>
              </button>
            ))}
          </div>
          {selectedMunicipality ? <p className="mt-3 text-sm">Выбран: {selectedMunicipality.name} (ID {selectedMunicipality.id})</p> : null}
        </Card>

        <Card className="bg-white/75 p-5 shadow-none">
          <SectionHeading title="Спортивные разряды" description="Справочник разрядов и срок действия." />
          {permissions.has("rank_applications.review") ? (
            <form className="mb-4 grid gap-2 sm:grid-cols-[1fr_120px_auto]" onSubmit={createRank}><Input name="name" placeholder="Название" required /><Input name="ttl" type="number" min="0" placeholder="Лет" required /><Button disabled={pending !== null}>{pending === "rank" ? <LoaderCircle className="animate-spin" /> : <Medal />}Добавить</Button></form>
          ) : null}
          <div className="grid max-h-80 gap-2 overflow-auto">
            {data.ranks.map((item) => (
              <button key={item.id} type="button" className="rounded-xl bg-muted/60 p-3 text-left text-sm hover:bg-muted" onClick={() => void rankApi.get(item.id).then(setSelectedRank).catch((cause) => notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить разряд"))}>
                {item.name} <span className="text-muted-foreground">#{item.id} · {item.ttl ? `${item.ttl} г.` : "бессрочно"}</span>
              </button>
            ))}
          </div>
          {selectedRank ? <p className="mt-3 text-sm">{selectedRank.name}: создан {formatDate(selectedRank.created_at)}</p> : null}
        </Card>
      </div> : null}

      {section === "reviews" && permissions.has("anti_doping.review") ? (
        <Card className="bg-white/75 p-5 shadow-none">
          <SectionHeading title="Проверка антидопингового сертификата" description="Backend пока не предоставляет очередь сертификатов, поэтому укажите ID из реестра файлов." />
          <form className="grid gap-3 md:grid-cols-2 xl:grid-cols-5" onSubmit={reviewCertificate}>
            <Input name="certificate_id" type="number" min="1" placeholder="ID сертификата" required />
            <Select name="status" defaultValue="valid"><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="valid">Действителен</SelectItem><SelectItem value="expired">Истёк</SelectItem><SelectItem value="rejected">Отклонён</SelectItem></SelectContent></Select>
            <DatePicker id="certificate-completed-at" name="completed_at" />
            <Input name="rejection_reason" placeholder="Причина отклонения" />
            <Button disabled={pending !== null}>{pending === "certificate" ? <LoaderCircle className="animate-spin" /> : null}Сохранить</Button>
          </form>
        </Card>
      ) : null}

      {section === "audit" && permissions.has("roles.read") ? (
        <Card className="bg-white/75 p-5 shadow-none">
          <SectionHeading title="Журнал аудита" description="Последние действия административных пользователей." />
          <div className="grid max-h-[32rem] gap-2 overflow-auto">
            {data.audit.map((entry) => (
              <div key={entry.id} className="grid gap-1 rounded-xl bg-muted/60 p-3 text-sm md:grid-cols-[1fr_auto]">
                <p className="font-medium"><Activity className="mr-2 inline size-4" />{entry.event_type} · {entry.method} {entry.route}</p>
                <p className="text-xs text-muted-foreground">{formatDate(entry.occurred_at, true)} · HTTP {entry.status}</p>
                <p className="text-xs text-muted-foreground">Пользователь #{entry.actor_user_id} · {entry.entity_type}{entry.entity_id ? ` #${entry.entity_id}` : ""}</p>
              </div>
            ))}
          </div>
        </Card>
      ) : null}
    </div>
  );
}
