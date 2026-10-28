"use client";

import { useCallback, useMemo, useState } from "react";
import {
  Archive,
  Download,
  File,
  FileCode2,
  FileImage,
  FileSpreadsheet,
  FileText,
  FileType,
  LoaderCircle,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { fileApi, type UploadedFile } from "@/entities/file";
import { formatDate } from "@/features/workspace/common";
import { useNotifications } from "@/shared/lib/notifications";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { downloadBlob } from "@/shared/lib/download";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";

function fileVisual(file: UploadedFile): { Icon: LucideIcon; color: string } {
  const extension = file.original_filename.split(".").pop()?.toLocaleLowerCase("ru") ?? "";
  if (file.content_type.startsWith("image/") || ["jpg", "jpeg", "png", "gif", "svg", "webp"].includes(extension)) return { Icon: FileImage, color: "bg-purple-500/10 text-purple-600" };
  if (file.content_type.includes("spreadsheet") || ["xls", "xlsx", "csv"].includes(extension)) return { Icon: FileSpreadsheet, color: "bg-emerald-500/10 text-emerald-600" };
  if (file.content_type.includes("pdf") || extension === "pdf") return { Icon: FileType, color: "bg-red-500/10 text-red-600" };
  if (file.content_type.includes("word") || ["doc", "docx"].includes(extension)) return { Icon: FileText, color: "bg-blue-500/10 text-blue-600" };
  if (file.content_type.includes("zip") || file.content_type.includes("archive") || ["zip", "rar", "7z", "tar", "gz"].includes(extension)) return { Icon: Archive, color: "bg-amber-500/10 text-amber-600" };
  if (file.content_type.includes("json") || file.content_type.includes("javascript") || ["js", "ts", "json", "html", "css"].includes(extension)) return { Icon: FileCode2, color: "bg-slate-500/10 text-slate-700" };
  if (file.content_type.startsWith("text/") || ["txt", "md", "rtf"].includes(extension)) return { Icon: FileText, color: "bg-blue-500/10 text-blue-600" };
  return { Icon: File, color: "bg-slate-500/10 text-slate-600" };
}

function fileSize(size: number) {
  if (size < 1024) return `${size} Б`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} КБ`;
  return `${(size / (1024 * 1024)).toFixed(1)} МБ`;
}

export function FilesList() {
  const notify = useNotifications();
  const [downloadingId, setDownloadingId] = useState<number | null>(null);
  const [selected, setSelected] = useState<UploadedFile | null>(null);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const load = useCallback(() => fileApi.list({ limit: 100 }), []);
  const state = useAsyncData(load);

  const files = useMemo(() => {
    const query = search.trim().toLocaleLowerCase("ru");
    return (state.data ?? []).filter((file) => {
      const matchesSearch = !query || `${file.original_filename} ${file.purpose} ${file.content_type}`.toLocaleLowerCase("ru").includes(query);
      return matchesSearch && (statusFilter === "all" || file.status === statusFilter);
    });
  }, [search, state.data, statusFilter]);
  const selectedVisual = selected ? fileVisual(selected) : null;
  const SelectedIcon = selectedVisual?.Icon;

  async function download(fileId: number) {
    setDownloadingId(fileId);
    try {
      const file = await fileApi.download(fileId);
      downloadBlob(file.blob, file.filename);
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось скачать файл");
    } finally {
      setDownloadingId(null);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar
        search={search}
        onSearchChange={setSearch}
        searchPlaceholder="Найти файл"
        filterContent={<Select value={statusFilter} onValueChange={setStatusFilter}><SelectTrigger><SelectValue placeholder="Статус" /></SelectTrigger><SelectContent><SelectItem value="all">Все файлы</SelectItem><SelectItem value="uploaded">Загруженные</SelectItem><SelectItem value="attached">Прикреплённые</SelectItem><SelectItem value="delete_pending">Ожидают удаления</SelectItem></SelectContent></Select>}
      />

      <Dialog open={detailsOpen} onOpenChange={setDetailsOpen}>
        <DialogContent>
          <DialogHeader className="sr-only">
            <DialogTitle>{selected?.original_filename}</DialogTitle>
            <DialogDescription>Просмотр файла и скачивание</DialogDescription>
          </DialogHeader>
          {selected && selectedVisual && SelectedIcon ? <div className="grid justify-items-center gap-4 text-center">
            <div className={`flex size-20 items-center justify-center rounded-3xl ${selectedVisual.color}`}><SelectedIcon className="size-11" /></div>
            <div className="grid gap-1"><p className="max-w-full break-words font-medium">{selected.original_filename}</p><p className="text-sm text-muted-foreground">{formatDate(selected.created_at, true)} · {fileSize(selected.size_bytes)}</p></div>
            <DialogFooter className="w-full"><Button className="w-full" disabled={downloadingId !== null} onClick={() => void download(selected.id)}>{downloadingId === selected.id ? <LoaderCircle className="animate-spin" /> : <Download />}Скачать файл</Button></DialogFooter>
          </div> : null}
        </DialogContent>
      </Dialog>

      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !files.length ? <DataEmpty title={search || statusFilter !== "all" ? "Файлы не найдены" : "Файлов пока нет"} description={search || statusFilter !== "all" ? "Измените поисковый запрос или фильтр." : "Загруженные документы появятся здесь."} /> : null}

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5 lg:gap-2">
        {files.map((file) => {
          const visual = fileVisual(file);
          const Icon = visual.Icon;
          return <Card
            key={file.id}
            role="button"
            tabIndex={0}
            className="group cursor-pointer border-0 bg-transparent p-3 shadow-none transition-colors hover:bg-white/40 focus-visible:ring-2 focus-visible:ring-ring sm:p-4 lg:p-3"
            onClick={() => { setSelected(file); setDetailsOpen(true); }}
            onKeyDown={(event) => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); setSelected(file); setDetailsOpen(true); } }}
          >
            <div className="flex min-h-32 flex-col items-center justify-center gap-2 text-center sm:min-h-36 lg:min-h-32">
              <div className={`flex size-16 items-center justify-center rounded-2xl transition-transform group-hover:scale-105 lg:size-14 ${visual.color}`}><Icon className="size-9 lg:size-8" /></div>
              <p className="w-full truncate text-sm font-medium" title={file.original_filename}>{file.original_filename}</p>
            </div>
          </Card>;
        })}
      </div>
    </div>
  );
}
