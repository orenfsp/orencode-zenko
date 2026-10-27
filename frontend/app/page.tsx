import { BrandLogo } from "@/shared/ui/brand-logo";
import { Button } from "@/shared/ui/button";
import RibbonGlowDark from "@/widgets/auth-shell/ui/bg-effect-dark";
import Link from "next/link";

export default function HomePage() {
  return (
    <div className="flex flex-col">
      <header className="z-20 fixed top-0 left-0 w-dvw p-3">
        <BrandLogo className="absolute left-10 top-10" />
      </header>
      <div className="h-[92dvh] relative">
        <RibbonGlowDark />
        <div className="absolute top-[50%] w-full -translate-y-[50%] text-white flex flex-col gap-10 items-center px-10 text-center">
          <h1 className="font-mono text-[80px] leading-[1] font-[300]">
            Выстраивайте путь <br /> к новым достижениям
          </h1>
          <p className="opacity-60 font-[300]">
            Соревнования, рейтинг и вся спортивная история всегда доступны
            спортсмену и тренеру.
          </p>
          <div className="grid grid-cols-2 gap-2">
            <Button asChild className="w-full rounded-xl" size="lg">
              <Link href="/#!">Ближайшие соревнования</Link>
            </Button>
            <Button
              asChild
              variant="outline"
              className="w-full bg-transparent rounded-xl"
              size="lg"
            >
              <Link href="/login">Войти в аккаунт</Link>
            </Button>
          </div>
        </div>
      </div>
      <div className="z-20 h-dvh w-full bg-white rounded-[40px] -mt-10"></div>
    </div>
  );
}
