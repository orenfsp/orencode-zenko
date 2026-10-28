"use client";

import { useCallback, useState } from "react";
import { Download, FileText, LoaderCircle } from "lucide-react";

import { fileApi } from "@/entities/file";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

export function FilesList() {
  const notify = useNotifications();
  const [downloadingId, setDownloadingId] = useState<number | null>(null);
  const load = useCallback(() => fileApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);

  async function download(fileId: number) {
    setDownloadingId(fileId);
    try {
      const file = await fileApi.download(fileId);
      const url = URL.createObjectURL(file.blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = file.filename;
      anchor.click();
      URL.revokeObjectURL(url);
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось скачать файл");
    } finally {
      setDownloadingId(null);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading title="Файлы" description="Загруженные документы и их технический статус." />
      {state.loading ? <DataLoading rows={5} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? <DataEmpty title="Файлов пока нет" /> : null}
      <div className="grid gap-3">
        {state.data?.map((file) => (
          <Card key={file.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div className="flex min-w-0 gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><FileText className="size-5" /></span>
                <div className="min-w-0">
                  <p className="truncate font-medium">{file.original_filename}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{file.purpose} · {(file.size_bytes / 1024).toFixed(1)} КБ · {formatDate(file.created_at, true)}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={file.status} />
                <Button variant="outline" size="sm" disabled={downloadingId !== null} onClick={() => void download(file.id)}>
                  {downloadingId === file.id ? <LoaderCircle className="animate-spin" /> : <Download />}
                  Скачать
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
