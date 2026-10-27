import { BrandLogo } from "@/shared/ui/brand-logo";
import { Button } from "@/shared/ui/button";
import RibbonGlowDark from "@/widgets/auth-shell/ui/bg-effect-dark";
import { ChevronLeft, ChevronRight } from "lucide-react";
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
      <div className="z-20 w-full bg-white rounded-[40px] -mt-10 py-20 pb-30 flex flex-col gap-10">
        <h2 className="font-mono text-[64px] leading-[1] font-[300] px-10 ">
          Новости <br /> и события
        </h2>
        <div className="w-full flex gap-5 overflow-x-auto overflow-y-hidden px-10 scrollbar-none">
          <div className="shrink-0 w-[calc((100%-20px)/2)] h-120 bg-primary rounded-[20px] overflow-hidden relative p-[2px]">
            <img
              src="/assets/phone-mockup.png"
              className="h-full w-full object-cover"
              alt=""
            />
            <div className="absolute bottom-0 left-0 bg-primary px-5 pt-5 pb-10 text-white flex flex-col gap-2">
              <p className="opacity-60 font-[300]">09 сентября 2026</p>
              <h3 className="font-mono text-[24px] leading-[1] font-[300]">
                Завершился первый Кубок Дальнего Востока по спортивному
                программированию
              </h3>
            </div>
          </div>
          <div className="shrink-0 w-[calc((100%-20px)/2)] h-120 bg-primary rounded-[20px] overflow-hidden relative p-[2px]">
            <img
              src="/assets/phone-mockup.png"
              className="h-full w-full object-cover"
              alt=""
            />
            <div className="absolute bottom-0 left-0 bg-primary px-5 pt-5 pb-10 text-white flex flex-col gap-2">
              <p className="opacity-60 font-[300]">09 сентября 2026</p>
              <h3 className="font-mono text-[24px] leading-[1] font-[300]">
                Завершился первый Кубок Дальнего Востока по спортивному
                программированию
              </h3>
            </div>
          </div>
          <div className="shrink-0 w-[calc((100%-20px)/2)] h-120 bg-primary rounded-[20px] overflow-hidden relative p-[2px]">
            <img
              src="/assets/phone-mockup.png"
              className="h-full w-full object-cover"
              alt=""
            />
            <div className="absolute bottom-0 left-0 bg-primary px-5 pt-5 pb-10 text-white flex flex-col gap-2">
              <p className="opacity-60 font-[300]">09 сентября 2026</p>
              <h3 className="font-mono text-[24px] leading-[1] font-[300]">
                Завершился первый Кубок Дальнего Востока по спортивному
                программированию
              </h3>
            </div>
          </div>
          <div className="shrink-0 w-[calc((100%-20px)/2)] h-120 bg-primary rounded-[20px] overflow-hidden relative p-[2px]">
            <img
              src="/assets/phone-mockup.png"
              className="h-full w-full object-cover"
              alt=""
            />
            <div className="absolute bottom-0 left-0 bg-primary px-5 pt-5 pb-10 text-white flex flex-col gap-2">
              <p className="opacity-60 font-[300]">09 сентября 2026</p>
              <h3 className="font-mono text-[24px] leading-[1] font-[300]">
                Завершился первый Кубок Дальнего Востока по спортивному
                программированию
              </h3>
            </div>
          </div>
        </div>
        <div className="px-10 flex items-center gap-4 justify-center">
          <Button variant="ghost" size="icon-xs">
            <ChevronLeft className="size-5" />
          </Button>
          <div className="flex  gap-1">
            <div className="h-1 w-1 bg-black opacity-20 rounded-full"></div>
            <div className="h-1 w-12 bg-black opacity-20 rounded-full"></div>
            <div className="h-1 w-1 bg-black opacity-20 rounded-full"></div>
            <div className="h-1 w-1 bg-black opacity-20 rounded-full"></div>
            <div className="h-1 w-1 bg-black opacity-20 rounded-full"></div>
          </div>

          <Button variant="ghost" size="icon-xs">
            <ChevronRight className="size-5" />
          </Button>
        </div>
      </div>
      <div className="z-20 w-full bg-[#1B1C21] rounded-[40px] -mt-10 py-20 pb-30 flex flex-col items-center gap-10">
        <h2 className="font-mono text-[64px] text-white leading-[1] font-[300] px-10 text-center">
          Ближайшие <br /> соревнования
        </h2>
        <div className="flex gap-2">
          <Button variant="secondary">Все</Button>
          <Button
            variant="ghost"
            className="text-white hover:bg-white/20 hover:text-white"
          >
            Все
          </Button>
          <Button
            variant="ghost"
            className="text-white hover:bg-white/20 hover:text-white"
          >
            Продуктовое
          </Button>
          <Button
            variant="ghost"
            className="text-white hover:bg-white/20 hover:text-white"
          >
            Алгоритмы
          </Button>
          <Button
            variant="ghost"
            className="text-white hover:bg-white/20 hover:text-white"
          >
            Информационная безопасность
          </Button>
          <Button
            variant="ghost"
            className="text-white hover:bg-white/20 hover:text-white"
          >
            Робототехника
          </Button>
          <Button
            variant="ghost"
            className="text-white hover:bg-white/20 hover:text-white"
          >
            БАС
          </Button>
        </div>
        <div className="grid grid-cols-2 gap-5 w-full px-10">
          <div className=" bg-black text-white rounded-[20px] w-full">
            <div className="flex w-full justify-between">
              <span className="opacity-60">Алгоритмическое программирование</span>
              <span className="opacity-40">10-12 Ноября</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
