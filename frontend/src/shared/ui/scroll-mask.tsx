"use client";

import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type CSSProperties,
  type HTMLAttributes,
} from "react";

import { cn } from "@/shared/lib/utils";

export const scrollMaskYClass = "scroll-mask-y";
export const scrollMaskXClass = "scroll-mask-x";

export function ScrollMask({
  children,
  className,
  fade = 20,
  axis = "y",
  style,
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  fade?: number;
  axis?: "x" | "y";
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [edges, setEdges] = useState({ start: false, end: false });

  const update = useCallback(() => {
    const element = ref.current;
    if (!element) return;
    const position = axis === "y" ? element.scrollTop : element.scrollLeft;
    const viewport = axis === "y" ? element.clientHeight : element.clientWidth;
    const content = axis === "y" ? element.scrollHeight : element.scrollWidth;
    const max = Math.max(0, content - viewport);
    const next = { start: max > 1 && position > 1, end: max > 1 && position < max - 1 };
    setEdges((current) => current.start === next.start && current.end === next.end ? current : next);
  }, [axis]);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;
    update();
    const resizeObserver = new ResizeObserver(update);
    resizeObserver.observe(element);
    if (element.firstElementChild) resizeObserver.observe(element.firstElementChild);
    const mutationObserver = new MutationObserver(update);
    mutationObserver.observe(element, { childList: true, subtree: true, characterData: true });
    element.addEventListener("scroll", update, { passive: true });
    return () => {
      resizeObserver.disconnect();
      mutationObserver.disconnect();
      element.removeEventListener("scroll", update);
    };
  }, [update]);

  const direction = axis === "y" ? "to bottom" : "to right";
  const mask = !edges.start && !edges.end
    ? "none"
    : edges.start && edges.end
      ? `linear-gradient(${direction}, transparent 0, black ${fade}px, black calc(100% - ${fade}px), transparent 100%)`
      : edges.start
        ? `linear-gradient(${direction}, transparent 0, black ${fade}px, black 100%)`
        : `linear-gradient(${direction}, black 0, black calc(100% - ${fade}px), transparent 100%)`;

  return (
    <div
      ref={ref}
      data-scroll-mask={axis}
      data-mask-start={edges.start || undefined}
      data-mask-end={edges.end || undefined}
      className={cn("min-h-0", axis === "y" ? "overflow-y-auto" : "overflow-x-auto", className)}
      style={{
        ...style,
        maskImage: mask,
        WebkitMaskImage: mask,
        maskMode: "alpha",
      } as CSSProperties}
      {...props}
    >
      {children}
    </div>
  );
}

function bindScrollMask(element: HTMLElement, axis: "x" | "y") {
  const update = () => {
    const position = axis === "y" ? element.scrollTop : element.scrollLeft;
    const viewport = axis === "y" ? element.clientHeight : element.clientWidth;
    const content = axis === "y" ? element.scrollHeight : element.scrollWidth;
    const max = Math.max(0, content - viewport);
    const start = max > 1 && position > 1;
    const end = max > 1 && position < max - 1;
    const fade = axis === "y" ? 20 : 28;
    const direction = axis === "y" ? "to bottom" : "to right";
    const mask = !start && !end
      ? "none"
      : start && end
        ? `linear-gradient(${direction}, transparent 0, black ${fade}px, black calc(100% - ${fade}px), transparent 100%)`
        : start
          ? `linear-gradient(${direction}, transparent 0, black ${fade}px, black 100%)`
          : `linear-gradient(${direction}, black 0, black calc(100% - ${fade}px), transparent 100%)`;
    element.style.maskImage = mask;
    element.style.webkitMaskImage = mask;
  };
  const resizeObserver = new ResizeObserver(update);
  resizeObserver.observe(element);
  Array.from(element.children).forEach((child) => resizeObserver.observe(child));
  const mutationObserver = new MutationObserver(() => {
    resizeObserver.disconnect();
    resizeObserver.observe(element);
    Array.from(element.children).forEach((child) => resizeObserver.observe(child));
    update();
  });
  mutationObserver.observe(element, { childList: true, subtree: true, characterData: true });
  element.addEventListener("scroll", update, { passive: true });
  update();
  return () => {
    resizeObserver.disconnect();
    mutationObserver.disconnect();
    element.removeEventListener("scroll", update);
  };
}

export function ScrollMaskController() {
  useEffect(() => {
    const bindings = new Map<HTMLElement, () => void>();
    const sync = () => {
      document.querySelectorAll<HTMLElement>(`.${scrollMaskYClass}, .${scrollMaskXClass}`).forEach((element) => {
        if (!bindings.has(element)) bindings.set(element, bindScrollMask(element, element.classList.contains(scrollMaskXClass) ? "x" : "y"));
      });
      bindings.forEach((cleanup, element) => {
        if (!element.isConnected) {
          cleanup();
          bindings.delete(element);
        }
      });
    };
    sync();
    const observer = new MutationObserver(sync);
    observer.observe(document.body, { childList: true, subtree: true });
    return () => {
      observer.disconnect();
      bindings.forEach((cleanup) => cleanup());
    };
  }, []);
  return null;
}
