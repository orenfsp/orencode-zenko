"use client";

import { useRef, useState } from "react";
import { format, isValid, parseISO } from "date-fns";
import { ru } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";

import { Button } from "@/shared/ui/button";
import { Calendar } from "@/shared/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/shared/ui/popover";
import { cn } from "@/shared/lib/utils";

interface DatePickerProps {
  id: string;
  name: string;
  defaultValue?: string;
  disabled?: boolean;
  required?: boolean;
  allowFuture?: boolean;
  className?: string;
}

function dateFromValue(value?: string) {
  if (!value) return undefined;
  const parsed = parseISO(value);
  return isValid(parsed) ? parsed : undefined;
}

export function DatePicker({
  id,
  name,
  defaultValue,
  disabled,
  required,
  allowFuture = false,
  className,
}: DatePickerProps) {
  const [date, setDate] = useState<Date | undefined>(() =>
    dateFromValue(defaultValue),
  );
  const [invalid, setInvalid] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);

  return (
    <>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            ref={triggerRef}
            id={id}
            type="button"
            variant="outline"
            disabled={disabled}
            data-empty={!date}
            aria-required={required}
            aria-invalid={invalid}
            aria-describedby={invalid ? `${id}-error` : undefined}
            className={cn(
              "h-8 w-full justify-between px-2.5 text-left font-normal data-[empty=true]:text-muted-foreground",
              className,
            )}
          >
            {date ? format(date, "d MMMM yyyy", { locale: ru }) : "Выберите дату"}
            <CalendarIcon className="size-4 text-muted-foreground" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto overflow-hidden p-0" align="start">
          <Calendar
            mode="single"
            selected={date}
            onSelect={(selected) => { setDate(selected); if (selected) setInvalid(false); }}
            locale={ru}
            captionLayout="dropdown"
            startMonth={new Date(new Date().getFullYear() - 120, 0, 1)}
            endMonth={allowFuture ? new Date(new Date().getFullYear() + 20, 11, 31) : new Date()}
            disabled={allowFuture ? undefined : { after: new Date() }}
          />
        </PopoverContent>
      </Popover>
      <input
        type="text"
        className="sr-only"
        tabIndex={-1}
        aria-hidden="true"
        name={name}
        value={date ? format(date, "yyyy-MM-dd") : ""}
        required={required}
        onChange={() => {}}
        onInvalid={(event) => { event.preventDefault(); setInvalid(true); triggerRef.current?.focus(); }}
      />
      {invalid ? <p id={`${id}-error`} className="mt-1 text-xs text-destructive">Выберите дату</p> : null}
    </>
  );
}
