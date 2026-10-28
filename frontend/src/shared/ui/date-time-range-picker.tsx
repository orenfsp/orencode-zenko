"use client";

import { useRef, useState } from "react";
import { format, isValid, parseISO } from "date-fns";
import { ru } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";
import type { DateRange } from "react-day-picker";

import { useIsMobile } from "@/hooks/use-mobile";
import { cn } from "@/shared/lib/utils";
import { Button } from "@/shared/ui/button";
import { Calendar } from "@/shared/ui/calendar";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/shared/ui/popover";

function parseDate(value?: string) {
  if (!value) return undefined;
  const parsed = parseISO(value);
  return isValid(parsed) ? parsed : undefined;
}

function dateTimeValue(date: Date | undefined, time: string) {
  return date ? `${format(date, "yyyy-MM-dd")}T${time}` : "";
}

function displayDate(date: Date) {
  return format(
    date,
    `d MMM${date.getFullYear() !== new Date().getFullYear() ? " yyyy" : ""}`,
    { locale: ru },
  );
}

interface DateTimeRangePickerProps {
  id: string;
  startName: string;
  endName: string;
  startDefaultValue?: string;
  endDefaultValue?: string;
  defaultStartTime?: string;
  defaultEndTime?: string;
  required?: boolean;
  placeholder?: string;
  className?: string;
}

export function DateTimeRangePicker({
  id,
  startName,
  endName,
  startDefaultValue,
  endDefaultValue,
  defaultStartTime = "10:00",
  defaultEndTime = "18:00",
  required,
  placeholder = "Выберите начало и окончание",
  className,
}: DateTimeRangePickerProps) {
  const isMobile = useIsMobile();
  const startDefaultDate = parseDate(startDefaultValue);
  const endDefaultDate = parseDate(endDefaultValue);
  const [range, setRange] = useState<DateRange | undefined>(() => ({
    from: startDefaultDate,
    to: endDefaultDate,
  }));
  const [startTime, setStartTime] = useState(
    startDefaultDate ? format(startDefaultDate, "HH:mm") : defaultStartTime,
  );
  const [endTime, setEndTime] = useState(
    endDefaultDate ? format(endDefaultDate, "HH:mm") : defaultEndTime,
  );
  const [invalid, setInvalid] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);

  const label = range?.from
    ? range.to
      ? `${displayDate(range.from)}, ${startTime} — ${displayDate(range.to)}, ${endTime}`
      : `${displayDate(range.from)}, ${startTime} — выберите окончание`
    : placeholder;

  function selectRange(nextRange: DateRange | undefined) {
    setRange(nextRange);
    if (nextRange?.from && nextRange.to) setInvalid(false);
  }

  return (
    <>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            ref={triggerRef}
            id={id}
            type="button"
            variant="outline"
            data-empty={!range?.from}
            aria-required={required}
            aria-invalid={invalid}
            aria-describedby={invalid ? `${id}-error` : undefined}
            className={cn(
              "h-9 w-full justify-between overflow-hidden px-3 text-left font-normal data-[empty=true]:text-muted-foreground",
              className,
            )}
          >
            <span className="truncate">{label}</span>
            <CalendarIcon className="size-4 shrink-0 text-muted-foreground" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto max-w-[calc(100vw-2rem)] overflow-x-auto p-0" align="start">
          <Calendar
            mode="range"
            selected={range}
            onSelect={selectRange}
            numberOfMonths={isMobile ? 1 : 2}
            locale={ru}
            captionLayout="dropdown"
          />
          <div className="grid grid-cols-2 gap-3 border-t p-3">
            <div className="grid gap-1.5">
              <Label htmlFor={`${id}-start-time`}>Время начала</Label>
              <Input id={`${id}-start-time`} type="time" value={startTime} onChange={(event) => setStartTime(event.target.value)} />
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor={`${id}-end-time`}>Время окончания</Label>
              <Input id={`${id}-end-time`} type="time" value={endTime} onChange={(event) => setEndTime(event.target.value)} />
            </div>
          </div>
        </PopoverContent>
      </Popover>
      <input
        type="text"
        className="sr-only"
        tabIndex={-1}
        aria-hidden="true"
        name={startName}
        value={dateTimeValue(range?.from, startTime)}
        required={required}
        onChange={() => {}}
        onInvalid={(event) => {
          event.preventDefault();
          setInvalid(true);
          triggerRef.current?.focus();
        }}
      />
      <input
        type="text"
        className="sr-only"
        tabIndex={-1}
        aria-hidden="true"
        name={endName}
        value={dateTimeValue(range?.to, endTime)}
        required={required}
        onChange={() => {}}
        onInvalid={(event) => {
          event.preventDefault();
          setInvalid(true);
          triggerRef.current?.focus();
        }}
      />
      {invalid ? <p id={`${id}-error`} className="mt-1 text-xs text-destructive">Выберите даты начала и окончания</p> : null}
    </>
  );
}
