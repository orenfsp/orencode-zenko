"use client";

import { useRef, useState } from "react";
import { format, isValid, parseISO } from "date-fns";
import { ru } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";

import { Button } from "@/shared/ui/button";
import { Calendar } from "@/shared/ui/calendar";
import { Input } from "@/shared/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/shared/ui/popover";
import { cn } from "@/shared/lib/utils";

function parseDateTime(value?: string) {
  if (!value) return undefined;
  const date = parseISO(value);
  return isValid(date) ? date : undefined;
}

export function DateTimePicker({ id, name, required, defaultValue, className }: { id: string; name: string; required?: boolean; defaultValue?: string; className?: string }) {
  const initial = parseDateTime(defaultValue);
  const [date, setDate] = useState<Date | undefined>(initial);
  const [time, setTime] = useState(initial ? format(initial, "HH:mm") : "12:00");
  const [invalid, setInvalid] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const value = date ? `${format(date, "yyyy-MM-dd")}T${time}` : "";

  return (
    <div className={cn("grid grid-cols-[1fr_6.5rem] gap-2", className)}>
      <Popover>
        <PopoverTrigger asChild><Button ref={triggerRef} id={id} type="button" variant="outline" aria-required={required} aria-invalid={invalid} aria-describedby={invalid ? `${id}-error` : undefined} className={cn("h-9 justify-between px-3 text-left font-normal", !date && "text-muted-foreground")}>{date ? format(date, "d MMMM yyyy", { locale: ru }) : "Выберите дату"}<CalendarIcon className="size-4 text-muted-foreground" /></Button></PopoverTrigger>
        <PopoverContent className="w-auto overflow-hidden p-0" align="start"><Calendar mode="single" selected={date} onSelect={(selected) => { setDate(selected); if (selected) setInvalid(false); }} locale={ru} captionLayout="dropdown" /></PopoverContent>
      </Popover>
      <Input type="time" value={time} onChange={(event) => setTime(event.target.value)} aria-label="Время" />
      <input type="text" className="sr-only" tabIndex={-1} aria-hidden="true" name={name} value={value} required={required} onChange={() => {}} onInvalid={(event) => { event.preventDefault(); setInvalid(true); triggerRef.current?.focus(); }} />
      {invalid ? <p id={`${id}-error`} className="col-span-full text-xs text-destructive">Выберите дату</p> : null}
    </div>
  );
}
