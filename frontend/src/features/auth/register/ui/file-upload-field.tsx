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
  const [fileName, setFileName] = useState<string | null>(null);

  const templateLinkClassName =
    "w-fit text-left text-sm leading-5 text-muted-foreground underline underline-offset-2 hover:text-foreground";

  return (
    <fieldset className="min-w-0 max-w-full space-y-2 border-0 p-0">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        name={name}
        type="file"
        accept=".pdf,.jpg,.jpeg,.png,.xml,.docx"
        className="peer sr-only"
        required={required}
        onChange={(event) =>
          setFileName(event.currentTarget.files?.[0]?.name ?? null)
        }
      />
      <Label
        htmlFor={id}
        title={fileName ?? undefined}
        className="flex h-9 w-full min-w-0 max-w-full cursor-pointer items-center overflow-hidden rounded-[var(--radius-control)] border border-input bg-background px-3 font-normal transition-[color,background-color,border-color,box-shadow] hover:bg-accent peer-focus-visible:border-ring peer-focus-visible:ring-[3px] peer-focus-visible:ring-ring/50"
      >
        <span className={`block min-w-0 max-w-full flex-1 truncate text-sm ${fileName ? "text-foreground" : "text-muted-foreground"}`}>
          {fileName ?? "Выберите файл"}
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
