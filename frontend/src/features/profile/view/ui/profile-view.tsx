"use client";

import { useEffect, useState } from "react";
import { useApiReady } from "@/_app/providers/api-provider";
import {
  AtSign,
  Building2,
  CalendarDays,
  Check,
  Code2,
  Info,
  Mail,
  MapPin,
  Pencil,
  ShieldCheck,
  Trophy,
  UserRound,
} from "lucide-react";
import {
  profileApi,
  type OwnProfile,
  type PublicProfile,
} from "@/entities/profile";
import { rbacApi, type Role } from "@/entities/rbac";
import { readTokens, storeTokens } from "@/entities/session";
import { cn } from "@/shared/lib/utils";
import { useNotifications } from "@/shared/lib/notifications";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/shared/ui/dialog";
import { FormField } from "@/shared/ui/form-field";
import { Separator } from "@/shared/ui/separator";
import { Skeleton } from "@/shared/ui/skeleton";
import { Spinner } from "@/shared/ui/spinner";

type ProfileViewProps =
  | { mode: "own"; userId?: never }
  | { mode: "public"; userId: number };

function fullName(profile: PublicProfile) {
  return [profile.last_name, profile.first_name, profile.middle_name]
    .filter(Boolean)
    .join(" ");
}

function formatDate(value?: string) {
  if (!value) return "Не указано";
  const date = new Date(`${value.slice(0, 10)}T00:00:00Z`);
  const includeYear = date.getUTCFullYear() !== new Date().getFullYear();
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    ...(includeYear ? { year: "numeric" as const } : {}),
    timeZone: "UTC",
  }).format(date);
}

function InfoRow({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Mail;
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex min-w-0 items-start gap-3 py-3.5">
      <Icon className="mt-1 size-4 shrink-0 text-muted-foreground" />
      <span className="min-w-0 flex-1">
        <span className="block text-xs text-muted-foreground">{label}</span>
        <span className="mt-0.5 block break-words text-sm font-medium">
          {value || "Не указано"}
        </span>
      </span>
    </div>
  );
}

function ProfileMetric({
  label,
  value,
}: {
  label: string;
  value: React.ReactNode;
}) {
  return (
    <Card className="flex-row items-center gap-3 border-0 bg-muted/50 p-4">
      <span className="min-w-0">
        <span className="block text-lg leading-none font-semibold">{value}</span>
        <span className="mt-1.5 block text-xs text-muted-foreground">
          {label}
        </span>
      </span>
    </Card>
  );
}

export function ProfileView(props: ProfileViewProps) {
  const apiReady = useApiReady();
  const notify = useNotifications();
  const [profile, setProfile] = useState<OwnProfile | PublicProfile | null>(
    null,
  );
  const [loading, setLoading] = useState(true);
  const [editOpen, setEditOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [telegram, setTelegram] = useState("");
  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);
  const [selectedRoleIds, setSelectedRoleIds] = useState<number[]>([]);
  const [rolesLoading, setRolesLoading] = useState(props.mode === "own");
  const [rolesError, setRolesError] = useState<string | null>(null);

  useEffect(() => {
    if (!apiReady) return;
    let active = true;
    const request =
      props.mode === "own" ? profileApi.me() : profileApi.public(props.userId);
    request
      .then((result) => {
        if (active) {
          setProfile(result);
          setTelegram(
            "tg_username" in result && typeof result.tg_username === "string"
              ? result.tg_username
              : "",
          );
          setSelectedRoleIds(result.roles.map((role) => role.id));
        }
      })
      .catch((cause) => {
        if (active)
          notify.error(
            cause instanceof Error
              ? cause.message
              : "Не удалось загрузить профиль",
          );
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [apiReady, notify, props.mode, props.userId]);

  useEffect(() => {
    if (!apiReady || props.mode !== "own") return;
    let active = true;
    rbacApi
      .selfAssignableRoles()
      .then((roles) => {
        if (!active) return;
        setAvailableRoles(roles);
        setRolesError(null);
      })
      .catch((cause) => {
        if (!active) return;
        setRolesError(
          cause instanceof Error ? cause.message : "Не удалось загрузить роли",
        );
      })
      .finally(() => {
        if (active) setRolesLoading(false);
      });
    return () => {
      active = false;
    };
  }, [apiReady, props.mode]);

  function changeEditOpen(open: boolean) {
    if (open && profile && "email" in profile) {
      setTelegram(profile.tg_username ?? "");
      setSelectedRoleIds(profile.roles.map((role) => role.id));
    }
    setEditOpen(open);
  }

  async function saveProfile(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!profile || !("email" in profile)) return;

    const currentRoleIds = profile.roles.map((role) => role.id);
    const rolesChanged =
      currentRoleIds.length !== selectedRoleIds.length ||
      currentRoleIds.some((roleId) => !selectedRoleIds.includes(roleId));
    const rolesEditable = currentRoleIds.every((roleId) =>
      availableRoles.some((role) => role.id === roleId),
    );

    if (rolesChanged && selectedRoleIds.length === 0) {
      notify.warning("Выберите хотя бы одну роль");
      return;
    }
    if (rolesChanged && !rolesEditable) {
      notify.warning("Системные роли нельзя изменить самостоятельно");
      return;
    }

    setSaving(true);
    try {
      const tokens = rolesChanged ? readTokens() : null;
      if (rolesChanged && !tokens) {
        throw new Error("Сессия не найдена. Войдите снова");
      }

      const result = await profileApi.update({
        tg_username: telegram.trim(),
        ...(rolesChanged
          ? {
              role_ids: selectedRoleIds,
              refresh_token: tokens!.refresh_token,
            }
          : {}),
      });

      if (result.access_token && result.refresh_token) {
        storeTokens({
          access_token: result.access_token,
          refresh_token: result.refresh_token,
        });
      }

      setProfile((current) =>
        current && "email" in current
          ? {
              ...current,
              tg_username: result.tg_username ?? null,
              roles: rolesChanged
                ? selectedRoleIds.flatMap((roleId) => {
                    const role = availableRoles.find(
                      (item) => item.id === roleId,
                    );
                    return role
                      ? [{ id: role.id, code: role.code, name: role.name }]
                      : [];
                  })
                : current.roles,
            }
          : current,
      );
      setEditOpen(false);
      notify.success("Профиль обновлён");
    } catch (cause) {
      notify.error(
        cause instanceof Error ? cause.message : "Не удалось обновить профиль",
      );
    } finally {
      setSaving(false);
    }
  }

  if (loading)
    return (
      <div className="mx-auto w-full max-w-6xl pb-24 md:pb-0">
        <Skeleton className="h-[720px] rounded-2xl" />
      </div>
    );
  if (!profile)
    return (
      <Card className="mx-auto max-w-3xl p-8">
        <CardTitle>Профиль не найден</CardTitle>
        <CardDescription className="mt-2">
          Проверьте ссылку или попробуйте позже.
        </CardDescription>
      </Card>
    );

  const ownProfile = props.mode === "own" ? (profile as OwnProfile) : null;
  const rolesEditable = ownProfile
    ? ownProfile.roles.every((profileRole) =>
        availableRoles.some((role) => role.id === profileRole.id),
      )
    : false;
  return (
    <Dialog open={editOpen} onOpenChange={changeEditOpen}>
      <div className="mx-auto w-full max-w-6xl pb-24 md:pb-0">
        <Card className="border-white/70 bg-white/90 backdrop-blur-xl">
          <header className="relative p-5 sm:p-6">
            <div className="flex flex-col items-center gap-5 text-center sm:flex-row sm:items-center sm:text-left">
              <Avatar className="size-28 shrink-0 rounded-2xl sm:size-24">
                <AvatarImage
                  src="/assets/workspace-avatar.png"
                  alt={fullName(profile)}
                  className="rounded-[inherit] object-cover grayscale"
                />
                <AvatarFallback className="rounded-[inherit] bg-brand-soft text-primary">
                  <UserRound className="size-10" />
                </AvatarFallback>
              </Avatar>

              <div className="min-w-0 flex-1">
                <h2 className="break-words font-mono text-2xl leading-tight font-medium tracking-tight sm:text-3xl">
                  {fullName(profile)}
                </h2>
                <p className="mt-2 text-sm text-muted-foreground">
                  {profile.age} лет
                  {profile.municipality
                    ? ` · ${profile.municipality.name}`
                    : ""}
                </p>
                <div className="mt-3 flex flex-wrap justify-center gap-2 sm:justify-start">
                  {profile.roles.map((role) => (
                    <Badge
                      key={role.id}
                      variant="secondary"
                      className="px-3 py-1"
                    >
                      {role.name}
                    </Badge>
                  ))}
                </div>
              </div>

              {ownProfile ? (
                <DialogTrigger asChild>
                  <Button className="hidden shrink-0 md:inline-flex">
                    <Pencil />
                    Редактировать
                  </Button>
                </DialogTrigger>
              ) : null}
            </div>
          </header>

          <Separator />

          <div className="grid gap-3 p-4 sm:grid-cols-3 sm:p-6">
            <ProfileMetric
              label="Возраст"
              value={`${profile.age} лет`}
            />
            <ProfileMetric
              label="Роли"
              value={profile.roles.length}
            />
            <ProfileMetric
              label="Разряды"
              value={profile.ranks.length}
            />
          </div>

          <div className="grid gap-4 px-4 pb-4 sm:px-6 sm:pb-6">
            <Card className="bg-background/75 p-4 sm:p-5">
              <CardHeader className="p-0 pb-3">
                <CardTitle>О профиле</CardTitle>
                <CardDescription>
                  {ownProfile
                    ? "Личные и контактные данные"
                    : "Открытая информация пользователя"}
                </CardDescription>
              </CardHeader>
              <Separator />
              {ownProfile ? (
                <>
                  <InfoRow icon={Mail} label="Email" value={ownProfile.email} />
                  <Separator />
                  <InfoRow
                    icon={AtSign}
                    label="Telegram"
                    value={
                      ownProfile.tg_username
                        ? `@${ownProfile.tg_username}`
                        : null
                    }
                  />
                  <Separator />
                  <InfoRow
                    icon={CalendarDays}
                    label="Дата рождения"
                    value={formatDate(ownProfile.birth_date)}
                  />
                  <Separator />
                </>
              ) : (
                <>
                  <InfoRow
                    icon={CalendarDays}
                    label="Возраст"
                    value={`${profile.age} лет`}
                  />
                  <Separator />
                </>
              )}
              <InfoRow
                icon={Building2}
                label="Организация"
                value={profile.organization?.name}
              />
              <Separator />
              <InfoRow
                icon={MapPin}
                label="Муниципалитет"
                value={profile.municipality?.name}
              />
            </Card>

            <Card className="bg-background/75 p-4 sm:p-5">
              <CardHeader className="p-0 pb-4">
                <CardTitle className="flex items-center gap-2">
                  <Code2 className="size-4 text-primary" />
                  Дисциплины
                </CardTitle>
              </CardHeader>
              <div className="flex flex-wrap gap-2">
                {profile.disciplines.length ? (
                  profile.disciplines.map((item) => (
                    <Badge key={item.id} variant="outline" className="px-3 py-1">
                      {item.name}
                    </Badge>
                  ))
                ) : (
                  <CardDescription>Дисциплины пока не указаны</CardDescription>
                )}
              </div>
            </Card>

            <Card className="bg-background/75 p-4 sm:p-5">
              <CardHeader className="p-0 pb-4">
                <CardTitle className="flex items-center gap-2">
                  <Trophy className="size-4 text-primary" />
                  Спортивные разряды
                </CardTitle>
              </CardHeader>
              <div className="grid gap-3">
                {profile.ranks.length ? (
                  profile.ranks.map((rank) => (
                    <div
                      key={rank.id}
                      className="rounded-xl border bg-muted/40 p-4"
                    >
                      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                        <div className="min-w-0">
                          <p className="break-words text-sm font-medium">
                            {rank.rank_name}
                          </p>
                          <p className="mt-1 break-words text-xs text-muted-foreground">
                            {rank.discipline_name}
                          </p>
                        </div>
                        {rank.expires_at ? (
                          <Badge variant="secondary">
                            до {formatDate(rank.expires_at)}
                          </Badge>
                        ) : (
                          <Badge variant="secondary">Бессрочно</Badge>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <CardDescription>
                    Подтверждённых разрядов пока нет
                  </CardDescription>
                )}
              </div>
            </Card>

            {profile.judge_category ? (
              <Card className="bg-background/75 p-4 sm:p-5">
                <CardHeader className="p-0">
                  <CardTitle className="flex items-center gap-2">
                    <ShieldCheck className="size-4 text-primary" />
                    Судейская категория
                  </CardTitle>
                  <CardDescription>
                    {profile.judge_category.category_name}
                    {profile.judge_category.expires_at
                      ? ` · до ${formatDate(profile.judge_category.expires_at)}`
                      : ""}
                  </CardDescription>
                </CardHeader>
              </Card>
            ) : null}
          </div>
        </Card>

        {ownProfile ? (
          <div className="fixed inset-x-0 bottom-0 z-40 bg-linear-to-t from-white via-white/95 to-transparent px-4 pt-7 pb-[max(1rem,env(safe-area-inset-bottom))] md:hidden">
            <DialogTrigger asChild>
              <Button size="lg" className="h-12 w-full rounded-xl">
                Редактировать
              </Button>
            </DialogTrigger>
          </div>
        ) : null}
      </div>

      {ownProfile ? (
        <DialogContent className="max-h-[calc(100svh-2rem)] overflow-y-auto sm:max-w-xl">
          <form onSubmit={saveProfile} className="grid gap-5">
            <DialogHeader>
              <DialogTitle>Редактирование профиля</DialogTitle>
              <DialogDescription>
                Измените контактные данные и доступные вам роли.
              </DialogDescription>
            </DialogHeader>
            <FormField
              id="tg-username"
              label="Telegram username"
              value={telegram}
              onChange={(event) => setTelegram(event.target.value)}
              placeholder="username"
              autoComplete="off"
            />

            <fieldset className="grid gap-3 border-0 p-0">
              <div>
                <legend className="text-sm font-medium">Роли</legend>
                <p className="mt-1 text-xs text-muted-foreground">
                  Можно выбрать несколько ролей. Первая выбранная роль станет
                  активной.
                </p>
              </div>

              {rolesLoading ? (
                <div className="flex min-h-20 items-center justify-center rounded-xl border bg-muted/30">
                  <Spinner />
                  <span className="ml-2 text-sm text-muted-foreground">
                    Загружаем роли…
                  </span>
                </div>
              ) : rolesError ? (
                <Alert variant="destructive">
                  <Info />
                  <AlertTitle>Роли недоступны</AlertTitle>
                  <AlertDescription>{rolesError}</AlertDescription>
                </Alert>
              ) : !rolesEditable ? (
                <Alert>
                  <Info />
                  <AlertTitle>Есть системная роль</AlertTitle>
                  <AlertDescription>
                    Такой набор ролей нельзя изменить самостоятельно. Обратитесь
                    к администратору.
                  </AlertDescription>
                </Alert>
              ) : (
                <div className="grid gap-2 sm:grid-cols-2">
                  {availableRoles.map((role) => {
                    const selected = selectedRoleIds.includes(role.id);
                    return (
                      <Button
                        key={role.id}
                        type="button"
                        variant={selected ? "secondary" : "outline"}
                        className={cn(
                          "h-auto min-h-12 justify-between whitespace-normal px-4 py-3 text-left",
                          selected && "border-primary/40 text-primary",
                        )}
                        aria-pressed={selected}
                        onClick={() =>
                          setSelectedRoleIds((current) =>
                            selected
                              ? current.filter((roleId) => roleId !== role.id)
                              : [...current, role.id],
                          )
                        }
                      >
                        <span>{role.name}</span>
                        {selected ? <Check className="size-4" /> : null}
                      </Button>
                    );
                  })}
                </div>
              )}
            </fieldset>

            <Alert>
              <Info />
              <AlertTitle>Остальные данные</AlertTitle>
              <AlertDescription>
                ФИО, email, дата рождения, организация и муниципалитет пока
                доступны только для просмотра.
              </AlertDescription>
            </Alert>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => changeEditOpen(false)}
              >
                Отмена
              </Button>
              <Button
                type="submit"
                disabled={saving}
              >
                {saving ? <Spinner /> : null}
                {saving ? "Сохраняем…" : "Сохранить"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}
