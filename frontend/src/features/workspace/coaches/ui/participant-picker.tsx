"use client";

import { useEffect, useRef, useState } from "react";
import { Check, ChevronsUpDown } from "lucide-react";

import { userCoachApi, type CoachParticipant } from "@/entities/user-coach";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/shared/ui/popover";
import { ScrollMask } from "@/shared/ui/scroll-mask";

function fullName(user: CoachParticipant) {
  return [user.last_name, user.first_name, user.middle_name].filter(Boolean).join(" ");
}

export function ParticipantPicker({ value, onChange }: { value: CoachParticipant | null; onChange: (user: CoachParticipant) => void }) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [users, setUsers] = useState<CoachParticipant[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);
  const [revision, setRevision] = useState(0);
  const firstOption = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!open) return;
    let active = true;
    const timer = setTimeout(() => {
      void userCoachApi.participants({ search: search.trim() || undefined, limit: 30 })
        .then((page) => { if (active) setUsers(page.items); })
        .catch(() => { if (active) { setUsers([]); setError(true); } })
        .finally(() => { if (active) setLoading(false); });
    }, search.trim() ? 250 : 0);
    return () => { active = false; clearTimeout(timer); };
  }, [open, search, revision]);

  return (
    <Popover open={open} onOpenChange={(next) => { setOpen(next); if (next) { setLoading(true); setError(false); } }}>
      <PopoverTrigger asChild>
        <Button type="button" variant="outline" role="combobox" aria-expanded={open} aria-label="Выбрать спортсмена" className="min-h-11 w-full min-w-0 justify-between font-normal">
          <span className="truncate">{value ? fullName(value) : "Выберите спортсмена"}</span>
          <ChevronsUpDown aria-hidden="true" className="shrink-0 opacity-60" />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-(--radix-popover-trigger-width) min-w-64 p-2">
        <Input
          value={search}
          onChange={(event) => { setSearch(event.target.value); setLoading(true); setError(false); setUsers([]); }}
          onKeyDown={(event) => { if (event.key === "ArrowDown" && users.length) { event.preventDefault(); firstOption.current?.focus(); } }}
          placeholder="Поиск по ФИО…"
          aria-label="Поиск спортсмена по имени"
          maxLength={100}
          autoComplete="off"
        />
        <ScrollMask className="mt-2 max-h-60" aria-live="polite">
          {loading ? <p className="px-2 py-3 text-sm text-muted-foreground">Ищем спортсменов…</p> : error ? (
            <div className="px-2 py-3 text-sm text-destructive">Не удалось загрузить список. <Button type="button" variant="link" className="h-auto p-0" onClick={() => { setLoading(true); setError(false); setRevision((current) => current + 1); }}>Повторить</Button></div>
          ) : users.length ? users.map((user, index) => (
            <button
              key={user.id}
              ref={index === 0 ? firstOption : undefined}
              type="button"
              className="flex min-h-11 w-full items-center justify-between gap-2 rounded-lg px-2 text-left text-sm hover:bg-accent focus-visible:bg-accent focus-visible:outline-none"
              onClick={() => { onChange(user); setOpen(false); }}
            >
              <span className="min-w-0 truncate">{fullName(user)}</span>
              {value?.id === user.id ? <Check aria-hidden="true" className="shrink-0" /> : null}
            </button>
          )) : <p className="px-2 py-3 text-sm text-muted-foreground">Спортсмены не найдены. Уточните ФИО.</p>}
        </ScrollMask>
        {!search && users.length === 30 ? <p className="px-2 pt-2 text-xs text-muted-foreground">Показаны первые 30. Чтобы найти другого спортсмена, введите ФИО.</p> : null}
      </PopoverContent>
    </Popover>
  );
}
