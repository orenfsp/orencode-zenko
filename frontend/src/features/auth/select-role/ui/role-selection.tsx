"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { CircleCheck, CirclePlus } from "lucide";

import { chooseRoles, readTokens, storeTokens } from "@/entities/session";
import { rbacApi, type Role } from "@/entities/rbac";
import { profileApi } from "@/entities/profile";
import { useNotifications } from "@/shared/lib/notifications";
import { BrandLogo } from "@/shared/ui/brand-logo";
import { Button } from "@/shared/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card";
import { MorphIcon } from "@/shared/ui/morph-icon";
import { Spinner } from "@/shared/ui/spinner";
import { cn } from "@/shared/lib/utils";

type RoleKey = "participant" | "coach" | "judge";

const roles: Array<{
  key: RoleKey;
  title: string;
  description: string;
  image: string;
  imagePosition: string;
}> = [
  {
    key: "participant",
    title: "Спортсмен",
    description:
      "Участвуйте в соревнованиях, отслеживайте результаты, рейтинг и путь к новому разряду.",
    image: "/assets/role-athlete.png",
    imagePosition: "50% 70%",
  },
  {
    key: "coach",
    title: "Тренер",
    description:
      "Сопровождайте спортсменов, подавайте заявки и следите за их результатами и разрядами.",
    image: "/assets/role-coach.png",
    imagePosition: "center",
  },
  {
    key: "judge",
    title: "Судья",
    description:
      "Получайте назначения, работайте с результатами и следите за своей судейской категорией.",
    image: "/assets/role-judge.png",
    imagePosition: "50% 45%",
  },
];

export function RoleSelection() {
  const router = useRouter();
  const notify = useNotifications();
  const [selected, setSelected] = useState<RoleKey[]>([]);
  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);
  const [rolesLoading, setRolesLoading] = useState(true);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    let active = true;
    Promise.all([rbacApi.selfAssignableRoles(), profileApi.me()])
      .then(([result, profile]) => {
        if (!active) return;
        setAvailableRoles(result);
        setSelected(
          profile.roles
            .map((role) => role.code)
            .filter((code): code is RoleKey => roles.some((item) => item.key === code)),
        );
      })
      .catch((cause) => { if (active) notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить роли"); })
      .finally(() => { if (active) setRolesLoading(false); });
      
    return () => { active = false; };
  }, [notify]);

  async function handleContinue() {
    setPending(true);
    const selectedRoles = selected.map((code) => availableRoles.find((item) => item.code === code));
    const tokens = readTokens();

    try {
      if (!tokens) throw new Error("Сессия не найдена. Войдите снова");
      if (selected.length === 0) throw new Error("Выберите хотя бы одну роль");
      if (selectedRoles.some((role) => !role)) throw new Error("Одна из выбранных ролей сейчас недоступна");
      const nextTokens = await chooseRoles(selectedRoles.map((role) => role!.id), tokens);
      storeTokens(nextTokens);
      notify.success(selected.length > 1 ? "Роли сохранены" : "Роль сохранена");
      router.push("/workspace");
    } catch (cause) {
      notify.error(
        cause instanceof Error ? cause.message : "Не удалось выбрать роль",
      );
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="relative min-h-svh overflow-x-hidden bg-background px-6 py-10 sm:px-10">
      <BrandLogo className="absolute left-6 top-10 sm:left-10" />

      <section className="mx-auto flex min-h-[calc(100svh-80px)] w-full max-w-[1200px] flex-col items-center justify-center pt-14">
        <header className="mb-10 w-full min-w-0 space-y-6 text-center">
          <h1 className="font-mono text-2xl font-medium leading-8">Выберите роль:</h1>
          <p className="break-words text-sm leading-5 text-muted-foreground">
            Выберите одну или несколько ролей в соответствии с вашими целями.
            В будущем вы сможете изменить их в профиле.
          </p>
        </header>

        <section className="grid w-full min-w-0 gap-6 md:grid-cols-3">
          {roles.map((role) => {
            const isSelected = selected.includes(role.key);

            return (
              <Card
                key={role.key}
                className={cn(
                  "h-[420px] w-full min-w-0 gap-4 transition-colors",
                  isSelected && "border-primary",
                )}
              >
                <figure className="relative h-64 w-full shrink-0 overflow-hidden">
                  <Image
                    src={role.image}
                    alt={role.title}
                    fill
                    priority
                    sizes="(min-width: 768px) 33vw, 100vw"
                    className="object-cover"
                    style={{ objectPosition: role.imagePosition }}
                  />
                </figure>

                <CardHeader className="h-[68px] min-w-0 shrink-0 px-4">
                  <CardTitle>{role.title}</CardTitle>
                  <CardDescription>{role.description}</CardDescription>
                </CardHeader>

                <CardFooter className="h-16 shrink-0 border-t bg-muted p-4">
                  <Button
                    type="button"
                    size="sm"
                    variant={isSelected ? "ghost" : "default"}
                    className={cn(
                      "w-full",
                      isSelected && "justify-start bg-transparent hover:bg-transparent",
                    )}
                    onClick={() => setSelected((current) => current.includes(role.key) ? current.filter((key) => key !== role.key) : [...current, role.key])}
                  >
                    {isSelected ? "Выбрана" : "Выбрать"}
                    <MorphIcon
                      icon={isSelected ? CircleCheck : CirclePlus}
                      size={16}
                      strokeWidth={2}
                      reducedMotion="user"
                    />
                  </Button>
                </CardFooter>
              </Card>
            );
          })}
        </section>

        <Button
          type="button"
          className="mt-10 px-5"
          onClick={handleContinue}
          disabled={pending || rolesLoading || selected.length === 0}
        >
          {pending || rolesLoading ? <Spinner /> : null}
          {rolesLoading ? "Загружаем роли…" : pending ? "Сохраняем…" : "Продолжить"}
        </Button>
      </section>
    </main>
  );
}
