"use client";

import { Filter, Plus, Search, X } from "lucide-react";
import type { ReactNode } from "react";
import { createPortal } from "react-dom";

import { Button } from "@/shared/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";

export function WorkspaceToolbar({
  search,
  onSearchChange,
  searchPlaceholder = "Поиск…",
  filterContent,
  filterDescription = "Уточните параметры отображаемых данных.",
  createLabel,
  onCreate,
  createClassName,
  filterActive = false,
  onResetFilters,
}: {
  search?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  filterContent?: ReactNode;
  filterDescription?: string;
  createLabel?: string;
  onCreate?: () => void;
  createClassName?: string;
  filterActive?: boolean;
  onResetFilters?: () => void;
}) {
  const content = (
    <div className="mx-auto flex min-w-0 w-full max-w-6xl flex-row items-center gap-1 py-2">
      {onSearchChange ? <div className="relative min-w-0 flex-1"><Search className="pointer-events-none absolute top-1/2 left-3 z-10 size-4 -translate-y-1/2 text-foreground/70" /><Input value={search ?? ""} onChange={(event) => onSearchChange(event.target.value)} placeholder={searchPlaceholder} aria-label={searchPlaceholder} className="bg-background/90 pl-9 shadow-none backdrop-blur-sm" /></div> : null}
      <div className="flex shrink-0 gap-1">
        {filterContent ? <Dialog><DialogTrigger asChild><Button variant="outline" aria-label="Открыть фильтры" aria-pressed={filterActive} className={`relative size-9 px-0 sm:h-9 sm:w-auto sm:px-3 ${filterActive ? "border-primary bg-primary/10 text-primary" : ""}`}><Filter className="size-4" /><span className="hidden sm:inline">Фильтры</span>{filterActive ? <span className="absolute right-1 top-1 size-1.5 rounded-full bg-primary" aria-hidden="true" /> : null}</Button></DialogTrigger><DialogContent><DialogHeader><DialogTitle>Фильтры</DialogTitle><DialogDescription>{filterDescription}</DialogDescription></DialogHeader><div className="grid gap-4">{filterContent}</div>{onResetFilters ? <Button type="button" variant="outline" className="w-full" onClick={onResetFilters}><X />Сбросить фильтры</Button> : null}</DialogContent></Dialog> : null}
        {onCreate ? <Button onClick={onCreate} aria-label={createLabel ?? "Создать"} className={`size-9 px-0 sm:h-9 sm:w-auto sm:px-3 ${createClassName ?? ""}`}><Plus className="size-4" /><span className="hidden sm:inline">{createLabel ?? "Создать"}</span></Button> : null}
      </div>
    </div>
  );

  const host = typeof document !== "undefined" ? document.querySelector<HTMLElement>("[data-workspace-toolbar-host]") : null;
  return host ? createPortal(content, host) : null;
}
