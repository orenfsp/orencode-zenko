"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import {
  accessTokenHasRole,
  getAccessIdentity,
  login,
  sessionMetadataStorage,
  storeTokens,
} from "@/entities/session";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { FormField } from "@/shared/ui/form-field";
import { Spinner } from "@/shared/ui/spinner";
import { FormHeading } from "@/widgets/auth-shell";

export function LoginForm() {
  const router = useRouter();
  const notify = useNotifications();
  const [pending, setPending] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);

    const data = new FormData(event.currentTarget);

    try {
      const tokens = await login(
        String(data.get("email") ?? ""),
        String(data.get("password") ?? ""),
      );
      storeTokens(tokens);
      const identity = getAccessIdentity(tokens.access_token);
      const email = String(data.get("email") ?? "").trim();
      sessionMetadataStorage.set({
        displayName: email.split("@")[0] || "Пользователь",
        email,
        roleIds: identity?.roleId ? [identity.roleId] : [],
      });
      notify.success("Вы вошли в аккаунт");
      router.push(
        accessTokenHasRole(tokens.access_token)
          ? process.env.NEXT_PUBLIC_AFTER_AUTH_URL || "/workspace"
          : "/choose-role",
      );
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось войти");
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="space-y-6" onSubmit={handleSubmit}>
      <FormHeading
        title="Войдите в свой аккаунт"
        description="Введите свою почту для продолжения"
      />

      <FormField
        id="email"
        name="email"
        type="email"
        label="Почта"
        placeholder="m@example.ru"
        autoComplete="email"
        required
      />

      <FormField
        id="password"
        name="password"
        type="password"
        label="Пароль"
        autoComplete="current-password"
        minLength={8}
        required
        labelAddon={
          <Link className="text-sm leading-5 hover:underline" href="/reset-password">
            Забыли пароль?
          </Link>
        }
      />

      <Button className="w-full" type="submit" disabled={pending}>
        {pending ? <Spinner /> : null}
        {pending ? "Входим…" : "Войти"}
      </Button>

      <p className="text-center text-sm leading-5">
        Нет аккаунта?{" "}
        <Link href="/register" className="underline underline-offset-2">
          Зарегистрироваться
        </Link>
      </p>
    </form>
  );
}
