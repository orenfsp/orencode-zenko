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
    <main className="grid h-svh grid-cols-2 overflow-hidden bg-card">
      <section className="relative flex h-svh min-h-0 items-center justify-center overflow-hidden px-10 py-16">
        <BrandLogo className="absolute left-10 top-10" />
        <section
          className={cn("w-full max-w-80 translate-y-5", contentClassName)}
        >
          {children}
        </section>
      </section>
      <aside
        className="relative h-svh w-full overflow-hidden bg-[#d9d9d9]"
        aria-hidden="true"
      >
        {/* <Image
          src="/assets/register_cover.png"
          alt=""
          fill
          priority
          sizes="50vw"
          className="object-cover object-top"
        /> */}
        <RibbonGlowDark />
        <div className="absolute top-0 left-0 w-full h-full p-5 text-white flex flex-col gap-5">
          <span className="font-mono text-[64px] leading-[1] font-[300]">
            Выстраивайте <br /> путь к новым <br /> достижениям
          </span>
          <span className="opacity-60 font-[300] w-1/2">
            Соревнования, рейтинг и вся спортивная история всегда доступны
            спортсмену и тренеру.
          </span>
          <div className="flex-1 min-h-0 -mx-5 -mb-5">
            <img
              src="/assets/phone-mockup.png"
              alt=""
              className="w-full h-full object-cover object-top"
            />
          </div>
        </div>
      </aside>
    </main>
  );
}
