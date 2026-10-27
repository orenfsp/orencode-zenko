"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import {
  isMinor,
  register,
  splitFullName,
  storeTokens,
} from "@/entities/session";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { DatePicker } from "@/shared/ui/date-picker";
import { FormField } from "@/shared/ui/form-field";
import { Label } from "@/shared/ui/label";
import { OAuthSeparator } from "@/shared/ui/oauth-separator";
import { Spinner } from "@/shared/ui/spinner";
import { VkButton } from "@/shared/ui/vk-button";
import { FormHeading } from "@/widgets/auth-shell";

import { FileUploadField } from "./file-upload-field";

interface RegistrationValues {
  fullName: string;
  email: string;
  birthDate: string;
  password: string;
}

type StepTransition =
  | "idle"
  | "exit-left"
  | "exit-right"
  | "enter-left"
  | "enter-right";

const initialValues: RegistrationValues = {
  fullName: "",
  email: "",
  birthDate: "",
  password: "",
};

export function RegisterForm() {
  const router = useRouter();
  const notify = useNotifications();
  const transitionTimers = useRef<number[]>([]);
  const [step, setStep] = useState<1 | 2>(1);
  const [stepTransition, setStepTransition] =
    useState<StepTransition>("idle");
  const [values, setValues] = useState(initialValues);
  const [pending, setPending] = useState(false);

  const minor = useMemo(
    () => Boolean(values.birthDate && isMinor(values.birthDate)),
    [values.birthDate],
  );

  useEffect(() => {
    return () => {
      transitionTimers.current.forEach((timer) => window.clearTimeout(timer));
    };
  }, []);

  function moveToStep(nextStep: 1 | 2, direction: "forward" | "backward") {
    transitionTimers.current.forEach((timer) => window.clearTimeout(timer));
    transitionTimers.current = [];

    setStepTransition(direction === "forward" ? "exit-left" : "exit-right");

    const exitTimer = window.setTimeout(() => {
      setStep(nextStep);
      setStepTransition(direction === "forward" ? "enter-right" : "enter-left");

      const enterTimer = window.setTimeout(() => {
        setStepTransition("idle");
      }, 260);

      transitionTimers.current.push(enterTimer);
    }, 180);

    transitionTimers.current.push(exitTimer);
  }

  function handleIdentitySubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const password = String(data.get("password") ?? "");
    const passwordConfirmation = String(data.get("password_confirmation") ?? "");
    const nextValues = {
      fullName: String(data.get("full_name") ?? "").trim(),
      email: String(data.get("email") ?? "").trim(),
      birthDate: String(data.get("birth_date") ?? ""),
      password,
    };

    if (splitFullName(nextValues.fullName).lastName === "") {
      notify.error("Укажите имя и фамилию");
      return;
    }
    if (!nextValues.birthDate) {
      notify.error("Укажите дату рождения");
      return;
    }
    if (password !== passwordConfirmation) {
      notify.error("Пароли не совпадают");
      return;
    }

    setValues(nextValues);
    moveToStep(2, "forward");
  }

  async function handleDocumentsSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);

    const files = new FormData(event.currentTarget);
    const personalConsent = files.get("personal_data_consent_file");
    const parentalConsent = files.get("parental_consent_document_file");

    if (!(personalConsent instanceof File) || personalConsent.size === 0) {
      notify.error("Добавьте согласие на обработку персональных данных");
      setPending(false);
      return;
    }
    if (
      minor &&
      (!(parentalConsent instanceof File) || parentalConsent.size === 0)
    ) {
      notify.error("Для несовершеннолетнего требуется разрешение родителей");
      setPending(false);
      return;
    }

    const name = splitFullName(values.fullName);
    const body = new FormData();
    body.append("first_name", name.firstName);
    body.append("last_name", name.lastName);
    if (name.middleName) body.append("middle_name", name.middleName);
    body.append("birth_date", values.birthDate);
    body.append("email", values.email);
    body.append("password", values.password);
    body.append("personal_data_consent_file", personalConsent);
    if (minor && parentalConsent instanceof File) {
      body.append("parental_consent_document_file", parentalConsent);
    }

    try {
      const tokens = await register(body);
      storeTokens(tokens);
      notify.success("Аккаунт создан", {
        description: "Теперь выберите роль для своего профиля.",
      });
      router.push("/choose-role");
    } catch (cause) {
      notify.error(
        cause instanceof Error ? cause.message : "Не удалось создать аккаунт",
      );
    } finally {
      setPending(false);
    }
  }

  if (step === 2) {
    return (
      <form
        className="registration-step space-y-6"
        data-transition={stepTransition}
        onSubmit={handleDocumentsSubmit}
      >
        <FormHeading
          title="Осталось немного"
          description="Отправьте указанные файлы для создания аккаунта"
        />

        <FileUploadField
          id="personal-consent"
          name="personal_data_consent_file"
          label="Согласие на обработку данных"
          templateHref={process.env.NEXT_PUBLIC_PERSONAL_CONSENT_TEMPLATE_URL}
          required
        />
        {minor &&
        <FileUploadField
          id="parental-consent"
          name="parental_consent_document_file"
          label="Разрешение родителей"
          templateHref={process.env.NEXT_PUBLIC_PARENTAL_CONSENT_TEMPLATE_URL}
          required={minor}
        />}

        <section className="space-y-3">
          <Button className="w-full" type="submit" disabled={pending}>
            {pending ? <Spinner /> : null}
            {pending ? "Создаём…" : "Создать аккаунт"}
          </Button>
          <Button
            className="w-full"
            type="button"
            variant="secondary"
            disabled={pending || stepTransition !== "idle"}
            onClick={() => moveToStep(1, "backward")}
          >
            Вернуться назад
          </Button>
        </section>
      </form>
    );
  }

  return (
    <form
      className="registration-step space-y-4"
      data-transition={stepTransition}
      onSubmit={handleIdentitySubmit}
    >
      <FormHeading
        title="Создайте аккаунт"
        description="Заполните все поля, чтобы продолжить"
      />

      <FormField
        id="full-name"
        name="full_name"
        label="Полное имя"
        placeholder="Иван Иванов"
        autoComplete="name"
        defaultValue={values.fullName}
        required
      />
      <FormField
        id="register-email"
        name="email"
        type="email"
        label="Почта"
        placeholder="m@example.ru"
        autoComplete="email"
        className="h-8"
        defaultValue={values.email}
        required
      />
      <fieldset className="space-y-2 border-0 p-0">
        <Label htmlFor="birth-date">Дата рождения</Label>
        <DatePicker
          id="birth-date"
          name="birth_date"
          defaultValue={values.birthDate}
        />
      </fieldset>
      <FormField
        id="register-password"
        name="password"
        type="password"
        label="Пароль"
        autoComplete="new-password"
        minLength={8}
        className="h-8"
        defaultValue={values.password}
        required
      />
      <FormField
        id="password-confirmation"
        name="password_confirmation"
        type="password"
        label="Подтвердить пароль"
        autoComplete="new-password"
        minLength={8}
        className="h-8"
        defaultValue={values.password}
        required
      />

      <Button
        className="w-full"
        type="submit"
        disabled={stepTransition !== "idle"}
      >
        Далее
      </Button>

      <OAuthSeparator />
      <VkButton />

      <p className="text-center text-sm leading-5 text-muted-foreground">
        Есть аккаунт?{" "}
        <Link href="/login" className="underline underline-offset-2">
          Войти
        </Link>
      </p>
    </form>
  );
}
