import type { CSSProperties, ReactNode } from "react";

import { cn } from "@/shared/lib/utils";

export const scrollMaskYClass = "scroll-mask-y";
export const scrollMaskXClass = "scroll-mask-x";

export function ScrollMask({
  children,
  className,
  fade = 20,
}: {
  children: ReactNode;
  className?: string;
  fade?: number;
}) {
  const mask = `linear-gradient(to bottom, transparent 0, black ${fade}px, black calc(100% - ${fade}px), transparent 100%)`;

  return (
    <div
      className={cn("min-h-0 overflow-y-auto", scrollMaskYClass, className)}
      style={{
        maskImage: mask,
        WebkitMaskImage: mask,
        maskMode: "alpha",
        WebkitMaskComposite: "source-in",
      } as CSSProperties}
    >
      {children}
    </div>
  );
}
