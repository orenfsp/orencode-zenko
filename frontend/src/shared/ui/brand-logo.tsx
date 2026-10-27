import Image from "next/image";
import Link from "next/link";

import { cn } from "@/shared/lib/utils";

export function BrandLogo({ className, href = "/" }: { className?: string; href?: string; theme?: "light" | "dark" }) {
  return (
    <Link
      href={href}
      className={cn("inline-flex h-6 items-center gap-2", className)}
      aria-label="Федерация спортивного программирования"
    >
      <Image
        src="/assets/fsp-mark.svg"
        alt=""
        width={25}
        height={24}
        priority
      />
      <Image
        src="/assets/fsp-wordmark.svg"
        alt="ФСП"
        width={36}
        height={16}
        priority
      />
    </Link>
  );
}
