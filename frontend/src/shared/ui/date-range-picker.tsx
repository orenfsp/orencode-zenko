"use client";

import { useState } from "react";
import { format, isValid, parseISO } from "date-fns";
import { ru } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";
import type { DateRange } from "react-day-picker";

import { Button } from "@/shared/ui/button";
import { Calendar } from "@/shared/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/shared/ui/popover";
import { cn } from "@/shared/lib/utils";

function parseDate(value?: string) {
  if (!value) return undefined;
  const date = parseISO(value);
  return isValid(date) ? date : undefined;
}

interface DateRangePickerProps {
  startName: string;
  endName: string;
  startDefaultValue?: string;
  endDefaultValue?: string;
  required?: boolean;
  className?: string;
  placeholder?: string;
}

export function DateRangePicker({ startName, endName, startDefaultValue, endDefaultValue, required, className, placeholder = "Выберите период" }: DateRangePickerProps) {
  const [range, setRange] = useState<DateRange | undefined>(() => ({
    from: parseDate(startDefaultValue),
    to: parseDate(endDefaultValue),
  }));
  const label = range?.from
    ? range.to
      ? `${format(range.from, "d MMM yyyy", { locale: ru })} — ${format(range.to, "d MMM yyyy", { locale: ru })}`
      : format(range.from, "d MMM yyyy", { locale: ru })
    : placeholder;

  return (
    <>
      <Popover>
        <PopoverTrigger asChild>
          <Button type="button" variant="outline" data-empty={!range?.from} className={cn("h-9 w-full justify-between px-3 text-left font-normal data-[empty=true]:text-muted-foreground", className)}>
            {label}<CalendarIcon className="size-4 text-muted-foreground" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto overflow-hidden p-0" align="start">
          <Calendar mode="range" selected={range} onSelect={setRange} numberOfMonths={2} locale={ru} captionLayout="dropdown" />
        </PopoverContent>
      </Popover>
      <input type="hidden" name={startName} value={range?.from ? format(range.from, "yyyy-MM-dd") : ""} required={required} />
      <input type="hidden" name={endName} value={range?.to ? format(range.to, "yyyy-MM-dd") : ""} required={required} />
    </>
  );
}
