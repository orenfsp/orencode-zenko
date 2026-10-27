"use client";

import { useState } from "react";

import { useNotifications } from "@/shared/lib/notifications";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";

interface FileUploadFieldProps {
  id: string;
  name: string;
  label: string;
  required?: boolean;
  templateHref?: string;
}

export function FileUploadField({
  id,
  name,
  label,
  required,
  templateHref,
}: FileUploadFieldProps) {
  const notify = useNotifications();
  const [fileName, setFileName] = useState("Файл не выбран");

  const templateLinkClassName =
    "w-fit text-left text-sm leading-5 text-muted-foreground underline underline-offset-2 hover:text-foreground";

  return (
    <fieldset className="space-y-2 border-0 p-0">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        name={name}
        type="file"
        accept=".pdf,.jpg,.jpeg,.png,.xml,.docx"
        className="peer sr-only"
        required={required}
        onChange={(event) =>
          setFileName(event.currentTarget.files?.[0]?.name ?? "Файл не выбран")
        }
      />
      <Label
        htmlFor={id}
        className="flex h-8 w-full cursor-pointer items-center overflow-hidden rounded-[10px] border border-border bg-background px-2.5 font-normal transition-shadow hover:bg-accent peer-focus-visible:border-ring peer-focus-visible:ring-2 peer-focus-visible:ring-ring/25"
      >
        <span className="shrink-0 pr-1 text-sm font-medium text-foreground">
          Выберите файл
        </span>
        <span className="min-w-0 flex-1 truncate text-sm text-muted-foreground">
          {fileName}
        </span>
      </Label>
      {templateHref ? (
        <a className={templateLinkClassName} href={templateHref} download>
          Скачать шаблон
        </a>
      ) : (
        <button
          type="button"
          className={templateLinkClassName}
          onClick={() =>
            notify.warning("Шаблон пока недоступен", {
              description: "Ссылка появится после подключения файла шаблона.",
            })
          }
        >
          Скачать шаблон
        </button>
      )}
    </fieldset>
  );
}
