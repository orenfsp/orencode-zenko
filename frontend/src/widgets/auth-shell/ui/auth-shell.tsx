import { BrandLogo } from "@/shared/ui/brand-logo";
import { cn } from "@/shared/lib/utils";
import Image from "next/image";
import RibbonGlowDark from "./bg-effect-dark";

export function AuthShell({
  children,
  contentClassName,
}: {
  children: React.ReactNode;
  contentClassName?: string;
}) {
  return (
    <main className="grid h-svh grid-cols-1 min-[1016px]:grid-cols-2 overflow-hidden bg-card">
  <section className="relative flex h-svh min-h-0 items-center justify-center overflow-hidden px-5 py-16 min-[1016px]:px-10">
    <BrandLogo className="absolute left-5 top-10 min-[1016px]:left-10" />

    <section
      className={cn("w-full max-w-80 translate-y-5", contentClassName)}
    >
      {children}
    </section>
  </section>

  <aside
    className="relative hidden h-svh w-full overflow-hidden bg-[#d9d9d9] min-[1016px]:block"
    aria-hidden="true"
  >
    <RibbonGlowDark />

    <div className="absolute top-0 left-0 flex h-full w-full flex-col gap-5 p-5 text-white">
      <span className="font-mono text-[64px] leading-[1] font-[300]">
        Выстраивайте <br /> путь к новым <br /> достижениям
      </span>

      <span className="w-1/2 font-[300] opacity-60">
        Соревнования, рейтинг и вся спортивная история всегда доступны
        спортсмену и тренеру.
      </span>

      <div className="relative -mx-5 -mb-5 min-h-0 flex-1">
        <Image
          src="/assets/phone-mockup.png"
          alt=""
          fill
          sizes="50vw"
          className="h-full w-full object-cover object-top"
        />
      </div>
    </div>
  </aside>
</main>
  );
}
