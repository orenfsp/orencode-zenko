"use client";

import * as React from "react";
import Image from "next/image";
import Link from "next/link";
import {
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  Menu,
  CalendarDays,
  MapPin,
} from "lucide-react";

import { useApiReady } from "@/_app/providers/api-provider";
import { competitionApi, type CalendarCompetition } from "@/entities/competition";
import { BrandLogo } from "@/shared/ui/brand-logo";
import { Button } from "@/shared/ui/button";
import RibbonGlowDark from "@/widgets/auth-shell/ui/bg-effect-dark";
import { cn } from "@/shared/lib/utils";
import {
  Sheet,
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetClose,
} from "@/shared/ui/sheet";

// Демонстрационные данные.
// Позже их можно заменить ответами API.
const news = [
  {
    id: 1,
    date: "09 октября",
    title:
      "Пять лет Федерации спортивного программирования: Москва принимает финал Кубка 2026",
    image: "/assets/new_1.png",
    category: "Новость",
  },
  {
    id: 2,
    date: "29 июля",
    title: "Федерация спортивного программирования представила технологические разработки на заседании комиссии Госсовета",
    image: "/assets/new_2.png",
    category: "Новость",
  },
  {
    id: 3,
    date: "20 сентября",
    title: "Пять лет Федерации спортивного программирования: в Москве подвели итоги Кубка ФСП 2026",
    image: "/assets/new_3.png",
    category: "Новость",
  },
  {
    id: 4,
    date: "09 сентября",
    title: "Завершился первый Кубок Дальнего Востока по спортивному программированию",
    image: "/assets/new_4.png",
    category: "Новость",
  },
];

function competitionDate(start: string, end: string) {
  const format = (value: string) => {
    const date = new Date(value);
    const includeYear = date.getFullYear() !== new Date().getFullYear();
    return new Intl.DateTimeFormat("ru-RU", {
      day: "numeric",
      month: "short",
      ...(includeYear ? { year: "numeric" as const } : {}),
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  };
  return `${format(start)} — ${format(end)}`;
}

function daysWord(value: number) {
  const remainder = value % 10;
  const tens = value % 100;
  if (remainder === 1 && tens !== 11) return "день";
  if (remainder >= 2 && remainder <= 4 && (tens < 10 || tens >= 20)) return "дня";
  return "дней";
}

function registrationLabel(startsAt?: string, endsAt?: string) {
  if (!startsAt || !endsAt) return "Срок регистрации уточняется";
  const now = Date.now();
  const start = new Date(startsAt).getTime();
  const end = new Date(endsAt).getTime();
  if (!Number.isFinite(start) || !Number.isFinite(end)) return "Срок регистрации уточняется";
  if (now < start) {
    const days = Math.max(1, Math.ceil((start - now) / 86_400_000));
    return `Регистрация откроется через ${days} ${daysWord(days)}`;
  }
  if (now >= end) return "Регистрация завершена";
  const days = Math.max(1, Math.ceil((end - now) / 86_400_000));
  return `Осталось ${days} ${daysWord(days)} регистрации`;
}

const competitionFormats: Record<string, string> = {
  offline: "Очно",
  online: "Онлайн",
  mixed: "Смешанный формат",
};

function SectionHeading({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <h2
      className={cn(
        "font-mono text-[36px] font-light leading-[0.98] tracking-[-0.045em]",
        "sm:text-[48px] lg:text-[64px] xl:text-[76px]",
        className,
      )}
    >
      {children}
    </h2>
  );
}

function HomeHeader() {
  const [menuOpen, setMenuOpen] = React.useState(false);
  const [headerVisible, setHeaderVisible] = React.useState(true);
  const [isScrolled, setIsScrolled] = React.useState(false);

  const lastScrollY = React.useRef(0);
  const scrollDirectionDistance = React.useRef(0);

  React.useEffect(() => {
    lastScrollY.current = window.scrollY;

    const handleScroll = () => {
      const currentScrollY = Math.max(0, window.scrollY);
      const difference = currentScrollY - lastScrollY.current;

      setIsScrolled(currentScrollY > 40);

      // В самом верху хедер всегда отображается
      if (currentScrollY <= 80) {
        setHeaderVisible(true);
        scrollDirectionDistance.current = 0;
      } else if (!menuOpen) {
        // Игнорируем слишком маленькие движения
        if (Math.abs(difference) < 1) {
          lastScrollY.current = currentScrollY;
          return;
        }

        // Сброс накопления при смене направления
        if (
          Math.sign(difference) !== Math.sign(scrollDirectionDistance.current)
        ) {
          scrollDirectionDistance.current = 0;
        }

        scrollDirectionDistance.current += difference;

        // Скролл вниз — прячем хедер
        if (scrollDirectionDistance.current > 8) {
          setHeaderVisible(false);
          scrollDirectionDistance.current = 0;
        }

        // Скролл вверх — показываем хедер
        if (scrollDirectionDistance.current < -8) {
          setHeaderVisible(true);
          scrollDirectionDistance.current = 0;
        }
      }

      lastScrollY.current = currentScrollY;
    };

    window.addEventListener("scroll", handleScroll, { passive: true });

    return () => {
      window.removeEventListener("scroll", handleScroll);
    };
  }, [menuOpen]);

  return (
    <header
      className={cn(
        "fixed inset-x-0 top-0 z-50",
        "transition-transform duration-300 ease-[cubic-bezier(0.22,1,0.36,1)]",
        "will-change-transform",
        isScrolled ? "bg-[#111217]/85 backdrop-blur-xl" : "bg-transparent",
      )}
      style={{
        transform:
          headerVisible || menuOpen
            ? "translate3d(0,0,0)"
            : "translate3d(0,-100%,0)",
      }}
    >
      <div className="relative mx-auto grid w-full max-w-[1800px] grid-cols-[1fr_auto] items-center px-5 py-4 sm:px-8 lg:grid-cols-[1fr_auto_1fr] lg:px-10">
        {/* Логотип */}
        <BrandLogo
          theme="light"
          href="/"
          aria-label="На главную"
          className="relative z-10 w-fit"
        />

        {/* Навигация на десктопе */}
        <nav
          aria-label="Основная навигация"
          className="hidden items-center justify-center gap-8 text-sm text-white/80 lg:flex"
        >
          <a href="#news" className="transition-colors hover:text-white">
            Новости
          </a>
          <a
            href="#competitions"
            className="transition-colors hover:text-white"
          >
            Соревнования
          </a>
          <a href="#contacts" className="transition-colors hover:text-white">
            Контакты
          </a>
        </nav>

        {/* Авторизация на десктопе */}
        <div className="hidden justify-self-end lg:block">
          <Button asChild variant="secondary" className="rounded-xl">
            <Link href="/login">
              Личный кабинет
              <ArrowRight className="ml-2 size-4" />
            </Link>
          </Button>
        </div>

        {/* Мобильное боковое меню */}
        <div className="justify-self-end lg:hidden">
          <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
            <SheetTrigger asChild>
              <button
                type="button"
                aria-label="Открыть меню"
                className={cn(
                  "flex size-10 items-center justify-center",
                  "rounded-xl bg-transparent text-white",
                  "border-0 outline-none",
                  "transition-colors hover:bg-white/10",
                  "focus-visible:ring-0 focus-visible:outline-none",
                )}
              >
                <Menu className="size-6" />
              </button>
            </SheetTrigger>

            <SheetContent
              side="right"
              className="flex h-dvh w-[min(280px,100vw)] flex-col gap-0 border-l-0 bg-[#1B1C21] p-0 text-white sm:max-w-[400px] [&>button]:text-white [&>button]:opacity-70"
            >
              <SheetHeader className=" px-6 pb-6 pt-8 text-left">
                <SheetTitle className="font-mono text-2xl font-light text-white">
                  Меню
                </SheetTitle>
              </SheetHeader>

              {/* Ссылки на секции */}
              <nav
                aria-label="Мобильная навигация"
                className="flex flex-1 flex-col gap-1 overflow-y-auto scroll-mask-y px-3 py-6"
              >
                {[
                  { label: "Новости", href: "#news" },
                  { label: "Соревнования", href: "#competitions" },
                  { label: "Контакты", href: "#contacts" },
                ].map((item) => (
                  <SheetClose asChild key={item.href}>
                    <a
                      href={item.href}
                      className="flex min-h-14 items-center justify-between rounded-xl px-4 text-base transition-colors hover:bg-white/10"
                    >
                      {item.label}
                      <ChevronRight className="size-5 text-white/40" />
                    </a>
                  </SheetClose>
                ))}
              </nav>

              {/* Кнопка входа внизу */}
              <div className=" px-5 pb-[calc(24px+env(safe-area-inset-bottom))] pt-5">
                <SheetClose asChild>
                  <Button asChild size="lg" className="h-12 w-full rounded-xl">
                    <Link href="/login">
                      Личный кабинет
                      <ArrowRight className="ml-2 size-4" />
                    </Link>
                  </Button>
                </SheetClose>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  );
}

function HeroSection() {
  return (
    <section className="relative flex min-h-[650px] h-[100svh] max-h-[1100px] items-center justify-center overflow-hidden bg-[#111217]">
      <div className="pointer-events-none absolute inset-0">
        <RibbonGlowDark />
      </div>

      <div className="relative z-10 mx-auto flex w-full max-w-[1500px] flex-col items-center gap-7 px-5 pt-16 text-center text-white sm:gap-9 sm:px-8 lg:px-10">
        <h1 className="max-w-[1200px] font-mono text-[clamp(38px,7.2vw,104px)] font-light leading-[0.98] tracking-[-0.055em]">
          Выстраивайте путь
          <br className="hidden sm:block" /> к новым достижениям
        </h1>

        <p className="max-w-[550px] text-sm font-light leading-relaxed text-white/60 sm:text-base lg:text-lg">
          Соревнования, рейтинг и вся спортивная история всегда доступны
          спортсмену и тренеру.
        </p>

        <div className="flex w-full max-w-[480px] flex-col gap-3 pt-2 sm:flex-row">
          <Button
            asChild
            size="lg"
            className="h-12 w-full rounded-xl sm:flex-1"
          >
            <a href="#competitions">
              Ближайшие соревнования
              <ArrowRight className="ml-2 size-4" />
            </a>
          </Button>

          <Button
            asChild
            size="lg"
            variant="outline"
            className="h-12 w-full rounded-xl border-white/30 bg-transparent text-white hover:bg-white/10 hover:text-white sm:flex-1"
          >
            <Link href="/login">Войти в аккаунт</Link>
          </Button>
        </div>
      </div>

      <a
        href="#news"
        className="absolute bottom-16 left-1/2 z-10 hidden -translate-x-1/2 flex-col items-center gap-3 text-white/50 transition-colors hover:text-white sm:flex"
      >
        <span className="text-xs">Листайте вниз</span>
        <div className="h-12 w-px bg-gradient-to-b from-white/50 to-transparent" />
      </a>
    </section>
  );
}

function NewsSection() {
  const AUTOPLAY_DELAY = 10_000;

  const scrollRef = React.useRef<HTMLDivElement>(null);
  const [activeIndex, setActiveIndex] = React.useState(0);
  const [progressKey, setProgressKey] = React.useState(0);

  const activeIndexRef = React.useRef(0);
  const programmaticScrollRef = React.useRef(false);
  const programmaticScrollTimerRef = React.useRef<
    ReturnType<typeof setTimeout> | undefined
  >(undefined);

  const resetProgress = React.useCallback(() => {
    setProgressKey((value) => value + 1);
  }, []);

  // Переключение на указанную карточку
  const scrollToCard = React.useCallback(
    (index: number) => {
      const container = scrollRef.current;
      if (!container) return;

      const cards = container.querySelectorAll<HTMLElement>("[data-news-card]");

      const target = cards[index];
      if (!target) return;

      const scrollPadding =
        Number.parseFloat(getComputedStyle(container).scrollPaddingLeft) || 0;

      if (programmaticScrollTimerRef.current !== undefined) {
        clearTimeout(programmaticScrollTimerRef.current);
      }

      programmaticScrollRef.current = true;

      container.scrollTo({
        left: target.offsetLeft - container.offsetLeft - scrollPadding,
        behavior: "smooth",
      });

      activeIndexRef.current = index;
      setActiveIndex(index);
      resetProgress();

      // Защита от сброса таймера во время плавной прокрутки
      programmaticScrollTimerRef.current = setTimeout(() => {
        programmaticScrollRef.current = false;
      }, 700);
    },
    [resetProgress],
  );

  // Отслеживаем ручную прокрутку карточек
  const updateActiveIndex = React.useCallback(() => {
    if (programmaticScrollRef.current) return;

    const container = scrollRef.current;
    if (!container) return;

    const cards = Array.from(
      container.querySelectorAll<HTMLElement>("[data-news-card]"),
    );

    if (!cards.length) return;

    const scrollPadding =
      Number.parseFloat(getComputedStyle(container).scrollPaddingLeft) || 0;

    const index = cards.reduce(
      (closest, card, index) => {
        const distance = Math.abs(
          card.offsetLeft -
            container.offsetLeft -
            container.scrollLeft -
            scrollPadding,
        );

        return distance < closest.distance ? { index, distance } : closest;
      },
      { index: 0, distance: Infinity },
    ).index;

    if (index !== activeIndexRef.current) {
      activeIndexRef.current = index;
      setActiveIndex(index);
      resetProgress();
    }
  }, [resetProgress]);

  // Автоматическая смена новости каждые 10 секунд
  React.useEffect(() => {
    if (news.length <= 1) return;

    const timer = window.setTimeout(() => {
      const nextIndex = (activeIndex + 1) % news.length;
      scrollToCard(nextIndex);
    }, AUTOPLAY_DELAY);

    return () => window.clearTimeout(timer);
  }, [activeIndex, progressKey, scrollToCard]);

  // Очистка таймера прокрутки при размонтировании
  React.useEffect(() => {
    return () => {
      if (programmaticScrollTimerRef.current !== undefined) {
        clearTimeout(programmaticScrollTimerRef.current);
      }
    };
  }, []);

  return (
    <section
      id="news"
      className="relative z-20 -mt-9 scroll-mt-8 rounded-t-[32px] bg-white py-16 text-[#1B1C21] sm:rounded-t-[40px] sm:py-20 lg:py-28"
    >
      <div className="mx-auto max-w-[1800px]">
        {/* Заголовок */}
        <div className="mb-8 flex items-end justify-between gap-5 px-5 sm:mb-12 sm:px-8 lg:px-10">
          <SectionHeading>
            Новости
            <br /> и события
          </SectionHeading>

          <p className="hidden max-w-[280px] text-sm leading-relaxed text-black/50 md:block">
            Следите за событиями, новостями и достижениями в спортивном
            программировании.
          </p>
        </div>

        {/* Горизонтальная карусель */}
        <div
          ref={scrollRef}
          onScroll={updateActiveIndex}
          className={cn(
            "scrollbar-none flex snap-x snap-mandatory",
            "gap-3 overflow-x-auto scroll-mask-x pb-2 sm:gap-5",
            "pl-5 pr-5 sm:pl-8 sm:pr-8 lg:pl-10 lg:pr-10",
            "scroll-pl-5 sm:scroll-pl-8 lg:scroll-pl-10",
            "scroll-pr-5 sm:scroll-pr-8 lg:scroll-pr-10",
          )}
        >
          {news.map((item) => (
            <article
              key={item.id}
              data-news-card
              className="group relative h-[390px] w-[85vw] max-w-[450px] shrink-0 snap-start overflow-hidden rounded-[20px] bg-[#1B1C21] sm:h-[460px] sm:w-[65vw] lg:h-[510px] lg:w-[calc((100%-20px)/2)] lg:max-w-none"
            >
              <Image
                src={item.image}
                alt=""
                fill
                sizes="(max-width: 640px) 85vw, (max-width: 1024px) 65vw, 50vw"
                className="object-cover object-center opacity-80 transition-transform duration-700 group-hover:scale-105"
              />

              <div className="absolute inset-0 bg-gradient-to-t from-black via-black/35 to-transparent" />

              <div className="absolute inset-x-0 bottom-0 flex flex-col gap-4 p-5 text-white sm:p-7 lg:p-8">
                <div className="flex flex-wrap items-center gap-3 text-xs text-white/65 sm:text-sm">
                  <span>{item.date}</span>
                  <span className="size-1 rounded-full bg-white/50" />
                  <span>{item.category}</span>
                </div>

                <h3 className="max-w-[600px] font-mono text-[23px] font-light leading-[1.08] tracking-tight sm:text-[28px] lg:text-[34px]">
                  {item.title}
                </h3>
              </div>
            </article>
          ))}
        </div>

        {/* Навигация и прогресс */}
        <div className="mt-8 flex items-center justify-between px-5 sm:mt-10 sm:px-8 lg:px-10">
          {/* Номер текущей новости */}
          <span className="text-sm text-black/40">
            {String(activeIndex + 1).padStart(2, "0")}
            <span className="mx-2">/</span>
            {String(news.length).padStart(2, "0")}
          </span>

          {/* Индикаторы прогресса */}
          <div className="flex items-center gap-1.5">
            {news.map((item, index) => (
              <button
                key={item.id}
                type="button"
                aria-label={`Перейти к новости ${index + 1}`}
                aria-current={index === activeIndex ? "true" : undefined}
                onClick={() => scrollToCard(index)}
                className={cn(
                  "relative h-1.5 overflow-hidden rounded-full",
                  "transition-[width] duration-300",
                  index === activeIndex
                    ? "w-12 bg-black/15"
                    : "w-1.5 bg-black/20 hover:bg-black/40",
                )}
              >
                {index === activeIndex && (
                  <span
                    key={`${index}-${progressKey}`}
                    className="absolute inset-y-0 left-0 w-full origin-left rounded-full bg-[#1B1C21] animate-[news-progress_10s_linear_forwards]"
                  />
                )}
              </button>
            ))}
          </div>

          {/* Стрелки навигации */}
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              size="icon"
              className="rounded-full"
              aria-label="Предыдущая новость"
              disabled={activeIndex === 0}
              onClick={() => scrollToCard(activeIndex - 1)}
            >
              <ChevronLeft className="size-5" />
            </Button>

            <Button
              type="button"
              variant="outline"
              size="icon"
              className="rounded-full"
              aria-label="Следующая новость"
              disabled={activeIndex === news.length - 1}
              onClick={() => scrollToCard(activeIndex + 1)}
            >
              <ChevronRight className="size-5" />
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}

function CompetitionCard({
  competition,
  registration,
}: {
  competition: CalendarCompetition;
  registration?: { startsAt: string; endsAt: string };
}) {
  return (
    <article className="flex min-h-[280px] flex-col justify-between gap-8 rounded-[20px] bg-[#111216] p-5 text-white sm:min-h-[310px] sm:p-7">
      <div className="flex flex-col gap-5">
        <div className="flex flex-wrap items-start justify-between gap-3 text-xs text-white/50 sm:text-sm">
          <span className="max-w-[260px]">{competition.discipline.name}</span>
          <span>{competitionDate(competition.start_at, competition.end_at)}</span>
        </div>

        <h3 className="max-w-[540px] font-mono text-[24px] font-light leading-[1.05] tracking-tight sm:text-[28px] xl:text-[32px]">
          {competition.name}
        </h3>
      </div>

      <div className="flex flex-col gap-5">
        <div className="flex flex-wrap gap-x-5 gap-y-2 text-xs text-white/45 sm:text-sm">
          <span className="flex items-center gap-2">
            <MapPin className="size-4" />
            {competition.venue_address || (competition.format === "online" ? "Онлайн-площадка" : "Место уточняется")}
          </span>
          <span className="flex items-center gap-2">
            <CalendarDays className="size-4" />
            {competitionFormats[competition.format] ?? competition.format}
          </span>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3">
          <Button asChild className="w-full rounded-xl sm:w-auto">
            <Link href="/login">
              Войти для участия
              <ArrowRight className="ml-2 size-4" />
            </Link>
          </Button>

          <span className="text-xs text-white/40">
            {registrationLabel(registration?.startsAt, registration?.endsAt)}
          </span>
        </div>
      </div>
    </article>
  );
}

function CompetitionsSection() {
  const apiReady = useApiReady();
  const [competitions, setCompetitions] = React.useState<CalendarCompetition[]>([]);
  const [registrations, setRegistrations] = React.useState<Record<number, { startsAt: string; endsAt: string }>>({});
  const [category, setCategory] = React.useState<number | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [revision, setRevision] = React.useState(0);

  React.useEffect(() => {
    if (!apiReady) return;
    let active = true;
    const from = new Date();
    const to = new Date(from);
    to.setFullYear(to.getFullYear() + 1);

    competitionApi.calendar({ from: from.toISOString(), to: to.toISOString(), limit: 100 })
      .then(async (items) => {
        if (active) {
          setCompetitions(items);
          setError(null);
        }
        const details = await Promise.allSettled(
          items.slice(0, 6).map(async (item) => {
            const detail = await competitionApi.get(item.id);
            return [item.id, { startsAt: detail.registration_starts_at, endsAt: detail.registration_ends_at }] as const;
          }),
        );
        if (active) {
          setRegistrations(Object.fromEntries(
            details.flatMap((result) => result.status === "fulfilled" ? [result.value] : []),
          ));
        }
      })
      .catch((cause) => {
        if (active) setError(cause instanceof Error ? cause.message : "Не удалось загрузить соревнования");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => { active = false; };
  }, [apiReady, revision]);

  const categories = React.useMemo(() => [
    { id: null, label: "Все" },
    ...Array.from(
      new Map(competitions.map((item) => [item.discipline.id, item.discipline.name])).entries(),
      ([id, label]) => ({ id, label }),
    ),
  ], [competitions]);

  const filteredCompetitions = React.useMemo(
    () =>
      (category === null
        ? competitions
        : competitions.filter((competition) => competition.discipline.id === category)
      ).slice(0, 6),
    [category, competitions],
  );

  return (
    <section
      id="competitions"
      className="relative z-30 -mt-9 scroll-mt-8 rounded-t-[32px] bg-[#1B1C21] py-16 text-white sm:rounded-t-[40px] sm:py-20 lg:py-28"
    >
      <div className="mx-auto flex w-full max-w-[1800px] flex-col items-center">
        <div className="px-5 text-center">
          <SectionHeading>
            Ближайшие
            <br /> соревнования
          </SectionHeading>
          <p className="mx-auto mt-6 max-w-[480px] text-sm leading-relaxed text-white/50 sm:text-base">
            Выберите дисциплину и найдите соревнование, в котором хотите принять
            участие.
          </p>
        </div>

        <div
          className={cn(
            "scrollbar-none mt-10 flex w-full snap-x gap-2",
            "overflow-x-auto scroll-mask-x pb-3",
            "pl-5 pr-5 sm:pl-8 sm:pr-8 lg:pl-10 lg:pr-10",
            "scroll-pl-5 sm:scroll-pl-8 lg:scroll-pl-10",
            "lg:justify-center",
          )}
        >
          {categories.map((item) => (
            <Button
              key={item.id}
              type="button"
              variant={category === item.id ? "secondary" : "ghost"}
              aria-pressed={category === item.id}
              onClick={() => setCategory(item.id)}
              className={cn(
                "h-10 shrink-0 snap-start rounded-xl px-4",
                category !== item.id &&
                  "text-white/60 hover:bg-white/10 hover:text-white",
              )}
            >
              {item.label}
            </Button>
          ))}
        </div>

        <div className="mt-5 grid w-full grid-cols-1 gap-3 px-5 sm:mt-8 sm:grid-cols-2 sm:gap-5 sm:px-8 lg:px-10">
          {loading ? (
            <div className="col-span-full py-16 text-center text-sm text-white/50" role="status">Загружаем соревнования…</div>
          ) : error ? (
            <div className="col-span-full flex min-h-[250px] flex-col items-center justify-center gap-4 rounded-[20px] border border-white/10 px-5 text-center">
              <p className="text-sm text-white/70">{error}</p>
              <Button variant="secondary" onClick={() => { setLoading(true); setRevision((value) => value + 1); }}>Повторить</Button>
            </div>
          ) : filteredCompetitions.length > 0 ? (
              filteredCompetitions.map((competition) => (
              <CompetitionCard key={competition.id} competition={competition} registration={registrations[competition.id]} />
            ))
          ) : (
            <div className="col-span-full flex min-h-[250px] flex-col items-center justify-center gap-3 rounded-[20px] border border-white/10 text-center">
              <CalendarDays className="size-9 text-white/30" />
              <p className="text-lg">Пока нет соревнований</p>
              <p className="text-sm text-white/40">
                {category === null ? "Опубликованные соревнования появятся здесь." : "Попробуйте выбрать другую дисциплину."}
              </p>
            </div>
          )}
        </div>

        <Button
          asChild
          size="lg"
          className="mt-10 rounded-xl"
          variant="secondary"
        >
          <Link href="/login">
            Перейти в личный кабинет
            <ArrowRight className="ml-2 size-4" />
          </Link>
        </Button>
      </div>
    </section>
  );
}

function Footer() {
  return (
    <footer
      id="contacts"
      className="relative z-40 -mt-9 rounded-t-[32px] bg-black px-5 pb-8 pt-16 text-white sm:rounded-t-[40px] sm:px-8 sm:pt-20 lg:px-10 lg:pt-24"
    >
      <div className="mx-auto flex max-w-[1800px] flex-col gap-16">
        <div className="flex flex-col justify-between gap-12 lg:flex-row lg:gap-16">
          <div className="flex flex-col gap-6">
            <Image
              src="/assets/full_logo.svg"
              alt="Федерация спортивного программирования"
              width={397}
              height={80}
              className="h-auto w-full max-w-[270px] object-contain object-left sm:max-w-[397px]"
            />

            <p className="max-w-[350px] text-sm leading-relaxed text-white/40">
              Цифровая платформа для участников, тренеров и организаторов
              соревнований по спортивному программированию.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-10 sm:grid-cols-2 lg:gap-20">
            <div className="flex flex-col items-start gap-3 text-sm">
              <span className="mb-2 text-white/40">Контакты</span>
              <a
                href="mailto:info@fsp-russia.ru"
                className="transition-opacity hover:opacity-60"
              >
                info@fsp-russia.ru
              </a>
              <a
                href="tel:+74996780305"
                className="transition-opacity hover:opacity-60"
              >
                +7 (499) 678-03-05
              </a>
              <span className="max-w-[260px] leading-relaxed text-white/70">
                125047, г. Москва,
                <br />
                2-я Брестская, д. 8, этаж 9
              </span>
            </div>

            <div className="flex flex-col items-start gap-3 text-sm">
              <span className="mb-2 text-white/40">Пресс-служба</span>
              <a
                href="mailto:press@fsp-russia.ru"
                className="transition-opacity hover:opacity-60"
              >
                press@fsp-russia.ru
              </a>
              <span className="mt-3 text-white/40">Социальные сети</span>
              <span className="text-white/70">ВКонтакте</span>
              <span className="text-white/70">Telegram</span>
            </div>
          </div>
        </div>

        <div className="flex flex-col justify-between gap-5  pt-6 text-xs text-white/40 sm:flex-row sm:items-center">
          <span>
            © {new Date().getFullYear()} Федерация спортивного программирования
          </span>

          <div className="flex flex-wrap gap-x-6 gap-y-3">
            <span>Политика конфиденциальности</span>
            <span>Пользовательское соглашение</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

export default function HomePage() {
  return (
    <main className="min-h-svh w-full overflow-x-clip bg-black">
      <HomeHeader />
      <HeroSection />
      <NewsSection />
      <CompetitionsSection />
      <Footer />
    </main>
  );
}
