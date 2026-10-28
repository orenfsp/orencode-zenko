"use client";

import * as React from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { cn } from "@/shared/lib/utils";

export interface ChartPoint {
  label: string;
  value: number;
}

export type ChartConfig = Record<string, { label?: React.ReactNode; color?: string }>;

const ChartContext = React.createContext<ChartConfig>({});

export function ChartContainer({
  config,
  children,
  className,
  label,
}: {
  config: ChartConfig;
  children: React.ReactElement;
  className?: string;
  label: string;
}) {
  const variables = Object.entries(config).reduce<Record<string, string>>((styles, [key, item]) => {
    if (item.color) styles[`--color-${key}`] = item.color;
    return styles;
  }, {});

  return (
    <ChartContext.Provider value={config}>
      <div
        role="img"
        aria-label={label}
        data-slot="chart"
        className={cn("h-56 min-w-0 w-full select-none outline-none focus:outline-none focus-visible:outline-none focus-visible:ring-0 [&_svg]:outline-none [&_svg:focus]:outline-none", className)}
        style={variables as React.CSSProperties}
      >
        <ResponsiveContainer width="100%" height="100%">
          {children}
        </ResponsiveContainer>
      </div>
    </ChartContext.Provider>
  );
}

export function ChartTooltipContent({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: Array<{ value?: number | string; name?: string; color?: string }>;
  label?: string;
}) {
  if (!active || !payload?.length) return null;
  return (
    <div className="grid min-w-28 gap-1 rounded-xl border bg-popover px-3 py-2 text-xs text-popover-foreground shadow-md">
      {label ? <p className="font-medium">{label}</p> : null}
      {payload.map((item, index) => <p key={`${item.name ?? "value"}-${index}`} className="flex items-center justify-between gap-3"><span className="text-muted-foreground">{item.name ?? "Значение"}</span><span className="font-medium tabular-nums">{item.value}</span></p>)}
    </div>
  );
}

export function ChartLine({ points, label }: { points: ChartPoint[]; label: string }) {
  if (!points.length) return null;
  const config = { value: { label: "Значение", color: "var(--primary)" } } satisfies ChartConfig;
  return (
    <ChartContainer config={config} label={label}>
      <LineChart accessibilityLayer={false} data={points} margin={{ top: 12, right: 12, left: 0, bottom: 8 }}>
        <CartesianGrid vertical={false} strokeDasharray="3 5" />
        <XAxis dataKey="label" tickLine={false} axisLine={false} tickMargin={8} tick={{ fontSize: 10 }} />
        <YAxis hide />
        <Tooltip cursor={false} content={<ChartTooltipContent />} />
        <Line type="monotone" dataKey="value" name="Значение" stroke="var(--color-value)" strokeWidth={3} dot={{ r: 4, fill: "var(--color-value)" }} activeDot={{ r: 5 }} />
      </LineChart>
    </ChartContainer>
  );
}

export function ChartBars({ points, label }: { points: ChartPoint[]; label: string }) {
  if (!points.length) return null;
  const config = { value: { label: "Баллы", color: "var(--primary)" } } satisfies ChartConfig;
  return (
    <ChartContainer config={config} label={label}>
      <BarChart accessibilityLayer={false} data={points} margin={{ top: 20, right: 8, left: 0, bottom: 8 }}>
        <CartesianGrid vertical={false} strokeDasharray="3 5" />
        <XAxis dataKey="label" tickLine={false} axisLine={false} tickMargin={8} tick={{ fontSize: 10 }} />
        <YAxis hide />
        <Tooltip cursor={false} content={<ChartTooltipContent />} />
        <Bar dataKey="value" name="Баллы" fill="var(--color-value)" radius={8} />
      </BarChart>
    </ChartContainer>
  );
}
