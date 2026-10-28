import { ArrowLeft } from "lucide-react";
import Image from "next/image";
import Link from "next/link";

import { AuthGuard } from "@/features/auth/auth-guard";
import { BrandLogo } from "@/shared/ui/brand-logo";
import { Button } from "@/shared/ui/button";

export function ResetPasswordPage() {
  return (
    <AuthGuard requireAuth={false} redirectTo="/workspace">
      <main className="relative flex min-h-svh items-center justify-center bg-white px-5 py-20">
        <BrandLogo className="absolute left-5 top-10 sm:left-10" />

        <section className="flex w-full max-w-md flex-col items-center gap-5 text-center">
          <Image
            src="/assets/reser-password.png"
            alt="Недоумевающий человек разводит руками"
            width={1024}
            height={1024}
            priority
            className="h-auto w-full max-w-sm rounded-[var(--radius-surface)] object-contain"
          />
          <p className="text-lg font-medium">Ну хз, этого не было в ТЗ</p>
          <Button asChild variant="outline">
            <Link href="/login">
              <ArrowLeft />
              Обратно
            </Link>
          </Button>
        </section>
      </main>
    </AuthGuard>
  );
}

