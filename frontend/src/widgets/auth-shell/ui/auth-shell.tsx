import { BrandLogo } from "@/shared/ui/brand-logo";
import { cn } from "@/shared/lib/utils";
import Image from "next/image";

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
        <Image
          src="/assets/login_cover.png"
          alt=""
          fill
          priority
          sizes="50vw"
          className="object-cover object-top"
        />
      </aside>
    </main>
  );
}
