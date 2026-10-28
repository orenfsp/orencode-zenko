"use client";

import * as React from "react";
import { CheckIcon } from "lucide-react";
import { Checkbox as CheckboxPrimitive } from "radix-ui";

import { cn } from "@/shared/lib/utils";

function Checkbox({ className, name, onCheckedChange, ...props }: React.ComponentProps<typeof CheckboxPrimitive.Root> & { name?: string }) {
  const [checked, setChecked] = React.useState<boolean | "indeterminate">(props.defaultChecked ?? false);
  const handleChange = (next: boolean | "indeterminate") => {
    setChecked(next);
    onCheckedChange?.(next);
  };
  return <>
    <CheckboxPrimitive.Root className={cn("peer size-4 shrink-0 rounded-md border border-input bg-background shadow-xs outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 data-[state=checked]:border-primary data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground disabled:cursor-not-allowed disabled:opacity-50", className)} onCheckedChange={handleChange} {...props}><CheckboxPrimitive.Indicator className="flex items-center justify-center"><CheckIcon className="size-3.5" /></CheckboxPrimitive.Indicator></CheckboxPrimitive.Root>
    {name ? <input type="hidden" name={name} value={checked === true ? "on" : ""} /> : null}
  </>;
}

export { Checkbox };
