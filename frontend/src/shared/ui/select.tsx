"use client";

import * as React from "react";
import { CheckIcon, ChevronDownIcon } from "lucide-react";
import { Select as SelectPrimitive } from "radix-ui";

import { cn } from "@/shared/lib/utils";

function Select({ name, required, defaultValue = "", onValueChange, ...props }: React.ComponentProps<typeof SelectPrimitive.Root> & { name?: string; required?: boolean }) {
  const [value, setValue] = React.useState(defaultValue);
  const handleChange = (nextValue: string) => {
    setValue(nextValue);
    onValueChange?.(nextValue);
  };

  return (
    <SelectPrimitive.Root name={name} required={required} value={value} onValueChange={handleChange} {...props} />
  );
}

function SelectTrigger({ className, children, ...props }: React.ComponentProps<typeof SelectPrimitive.Trigger>) {
  return (
    <SelectPrimitive.Trigger className={cn("flex h-9 w-full items-center justify-between gap-2 rounded-xl border border-input bg-background px-3 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50", className)} {...props}>
      {children}
      <SelectPrimitive.Icon><ChevronDownIcon className="size-4 text-muted-foreground" /></SelectPrimitive.Icon>
    </SelectPrimitive.Trigger>
  );
}

function SelectValue(props: React.ComponentProps<typeof SelectPrimitive.Value>) {
  return <SelectPrimitive.Value {...props} />;
}

function SelectContent({ className, ...props }: React.ComponentProps<typeof SelectPrimitive.Content>) {
  return (
    <SelectPrimitive.Portal>
      <SelectPrimitive.Content position="popper" className={cn("z-50 max-h-72 min-w-(--radix-select-trigger-width) overflow-y-auto rounded-xl border bg-popover p-1 text-popover-foreground shadow-md", className)} {...props} />
    </SelectPrimitive.Portal>
  );
}

function SelectItem({ className, children, ...props }: React.ComponentProps<typeof SelectPrimitive.Item>) {
  return (
    <SelectPrimitive.Item className={cn("relative flex w-full cursor-default items-center rounded-lg py-1.5 pr-8 pl-2 text-sm outline-none select-none focus:bg-accent focus:text-accent-foreground data-disabled:pointer-events-none data-disabled:opacity-50", className)} {...props}>
      <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
      <SelectPrimitive.ItemIndicator className="absolute right-2"><CheckIcon className="size-4" /></SelectPrimitive.ItemIndicator>
    </SelectPrimitive.Item>
  );
}

export { Select, SelectContent, SelectItem, SelectTrigger, SelectValue };
