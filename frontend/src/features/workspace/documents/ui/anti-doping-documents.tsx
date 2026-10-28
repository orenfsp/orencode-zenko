"use client";

import { ChangeEvent, useCallback, useState } from "react";
import { FileCheck2, LoaderCircle, Upload } from "lucide-react";

import { antiDopingApi } from "@/entities/anti-doping";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Input } from "@/shared/ui/input";

export function AntiDopingDocuments() {
  const notify = useNotifications();
  const [uploading, setUploading] = useState(false);
  const load = useCallback(() => antiDopingApi.mine(), []);
  const state = useAsyncData(load);

  async function upload(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    setUploading(true);
    try {
      await antiDopingApi.upload(file);
      notify.success("Сертификат загружен");
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось загрузить сертификат");
    } finally {
      setUploading(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-5xl">
      <SectionHeading
        title="Антидопинг"
        description="Сертификаты и результаты проверки федерацией. Поддерживаются PDF, JPEG и PNG до 10 МБ."
        action={
          <Button asChild>
            <label aria-disabled={uploading}>
              {uploading ? <LoaderCircle className="animate-spin" /> : <Upload />}
              Загрузить
              <Input type="file" accept="application/pdf,image/jpeg,image/png" className="sr-only" disabled={uploading} onChange={(event) => void upload(event)} />
            </label>
          </Button>
        }
      />
      {state.loading ? <DataLoading rows={3} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !state.data?.length ? <DataEmpty title="Сертификатов пока нет" description="Загрузите первый сертификат для проверки." /> : null}
      <div className="grid gap-3">
        {state.data?.map((certificate) => (
          <Card key={certificate.id} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><FileCheck2 className="size-5" /></span>
                <div>
                  <p className="font-medium">Сертификат #{certificate.id}</p>
                  <p className="mt-1 text-sm text-muted-foreground">Загружен {formatDate(certificate.created_at, true)}</p>
                  {certificate.valid_until ? <p className="mt-1 text-xs text-muted-foreground">Действует до {formatDate(certificate.valid_until)}</p> : null}
                  {certificate.rejection_reason ? <p className="mt-2 text-sm text-destructive">{certificate.rejection_reason}</p> : null}
                </div>
              </div>
              <StatusBadge status={certificate.status} />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
