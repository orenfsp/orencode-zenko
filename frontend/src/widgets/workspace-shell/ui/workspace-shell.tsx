"use client";

import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type ReactNode,
} from "react";
import {
  Award,
  Bell,
  Building2,
  CalendarClock,
  CalendarRange,
  Check,
  ChevronRight,
  ChevronsUpDown,
  ClipboardList,
  Dumbbell,
  FileCheck2,
  FileText,
  Gavel,
  LayoutGrid,
  LoaderCircle,
  ListTree,
  LogOut,
  Menu,
  Medal,
  Star,
  ShieldCheck,
  Moon,
  Sun,
  Trophy,
  UsersRound,
  UserRound,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import {
  profileApi,
  type OwnProfile,
  type ProfileRole,
} from "@/entities/profile";
import {
  chooseRoles,
  readTokens,
  storeTokens,
  useAuth,
} from "@/entities/session";
import { useNotifications } from "@/shared/lib/notifications";
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar";
import { BrandLogo } from "@/shared/ui/brand-logo";
import { Button } from "@/shared/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
  useSidebar,
} from "@/shared/ui/sidebar";
import RibbonGlow from "@/widgets/auth-shell/ui/bg-effect";
import RibbonGlowDark from "@/widgets/auth-shell/ui/bg-effect-dark";
import { Skeleton } from "@/shared/ui/skeleton";
import { cn } from "@/shared/lib/utils";
import { ScrollMask } from "@/shared/ui/scroll-mask";
import { useTheme } from "@/_app/providers/theme-provider";

const workspacePageTitles: Array<[string, string]> = [
  ["/workspace/profile", "Профиль"],
  ["/workspace/competitions/", "Соревнование"],
  ["/workspace/competitions", "Соревнования"],
  ["/workspace/applications", "Заявки и решения"],
  ["/workspace/deadlines", "Важные сроки"],
  ["/workspace/coaches", "Тренеры и спортсмены"],
  ["/workspace/results", "Спортивные результаты"],
  ["/workspace/ranks", "Спортивный разряд"],
  ["/workspace/documents", "Антидопинговые документы"],
  ["/workspace/teams", "Сборные команды"],
  ["/workspace/judge-competitions", "Мои соревнования"],
  ["/workspace/notifications", "Уведомления"],
  ["/workspace/registry", "Реестр"],
  ["/workspace/calendar-plans", "Календарные планы"],
  ["/workspace/files", "Документы пользователей"],
  ["/workspace/rating", "Рейтинг спортсменов"],
  ["/workspace/administration/rank-standards", "Разряды и нормативы"],
  ["/workspace/administration/roles", "Роли и права"],
  ["/workspace/administration/directories", "Системные справочники"],
  ["/workspace/administration/reviews", "Проверка сертификатов"],
  ["/workspace/administration/audit", "Журнал аудита"],
];

function pageTitle(pathname: string) {
  return workspacePageTitles.find(([path]) => pathname === path || pathname.startsWith(path))?.[1] ?? "Главная";
}

interface NavigationItem {
  label: string;
  icon: LucideIcon;
  href: string;
}

interface NavigationGroup {
  label: string;
  items: NavigationItem[];
}

const roleIcons: Record<string, LucideIcon> = {
  participant: UserRound,
  coach: Dumbbell,
  judge: Gavel,
  federation: Building2,
  super_admin: ShieldCheck,
};

const selfAssignableRoleCodes = new Set(["participant", "coach", "judge"]);

function profileName(profile: OwnProfile | null) {
  if (!profile) return "Пользователь";
  return [profile.first_name, profile.last_name].filter(Boolean).join(" ");
}

function RoleIcon({ role }: { role: ProfileRole }) {
  const Icon = roleIcons[role.code] ?? ShieldCheck;
  return <Icon />;
}

function ProfileAvatar({
  name,
  className,
}: {
  name: string;
  className?: string;
}) {
  return (
    <Avatar className={className ?? "size-8 rounded-lg"}>
      <AvatarImage
        src="/assets/workspace-avatar.png"
        alt={name}
        className="rounded-[inherit] object-cover grayscale"
      />
      <AvatarFallback className="rounded-[inherit] bg-brand-soft text-primary">
        <UserRound className="size-4" />
      </AvatarFallback>
    </Avatar>
  );
}

function useWorkspaceUser() {
  const { identity, logout } = useAuth();
  const notify = useNotifications();
  const router = useRouter();

  const [profile, setProfile] = useState<OwnProfile | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [profileError, setProfileError] = useState(false);
  const [profileRevision, setProfileRevision] = useState(0);
  const [pendingRoleId, setPendingRoleId] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;

    queueMicrotask(() => {
      if (cancelled) return;
      setProfile(null);
      setProfileLoading(true);
      setProfileError(false);
    });

    if (!identity) {
      return () => {
        cancelled = true;
      };
    }

    const fetchProfile = async () => {
      try {
        const result = await profileApi.me();

        if (cancelled) return;

        setProfile(result);
        setProfileLoading(false);
        setProfileError(false);
      } catch {
        if (cancelled) return;
        setProfileLoading(false);
        setProfileError(true);

        retryTimer = setTimeout(() => {
          void fetchProfile();
        }, 3000);
      }
    };

    void fetchProfile();

    return () => {
      cancelled = true;

      if (retryTimer !== undefined) {
        clearTimeout(retryTimer);
      }
    };
  }, [identity, profileRevision]);

  const activeRole = useMemo(
    () => profile?.roles.find((role) => role.id === identity?.roleId) ?? null,
    [identity?.roleId, profile?.roles],
  );

  const rolesCanBeReordered = Boolean(
    profile?.roles.length &&
    profile.roles.every((role) => selfAssignableRoleCodes.has(role.code)),
  );

  async function switchRole(role: ProfileRole) {
    if (
      !profile ||
      role.id === identity?.roleId ||
      pendingRoleId !== null ||
      !rolesCanBeReordered
    ) {
      return;
    }

    const tokens = readTokens();

    if (!tokens) {
      notify.error("Сессия не найдена. Войдите снова");
      return;
    }

    setPendingRoleId(role.id);

    try {
      const orderedRoleIds = [
        role.id,
        ...profile.roles
          .filter((assignedRole) => assignedRole.id !== role.id)
          .map((assignedRole) => assignedRole.id),
      ];

      const nextTokens = await chooseRoles(orderedRoleIds, tokens);

      storeTokens(nextTokens);

      notify.success(`Активная роль: ${role.name}`);

      router.replace("/workspace");
      router.refresh();
    } catch (cause) {
      notify.error(
        cause instanceof Error ? cause.message : "Не удалось переключить роль",
      );
    } finally {
      setPendingRoleId(null);
    }
  }

  async function signOut() {
    try {
      await logout();
    } catch {
      // Локальные токены удаляются в logout даже при сетевой ошибке.
    }
    router.replace("/login");
  }

  return {
    identity,
    profile,
    profileLoading,
    profileError,
    retryProfile: () => setProfileRevision((value) => value + 1),
    pendingRoleId,
    activeRole,
    displayName: profileName(profile),
    rolesCanBeReordered,
    switchRole,
    signOut,
  };
}

type WorkspaceUser = ReturnType<typeof useWorkspaceUser>;

function workspaceNavigation(user: WorkspaceUser): NavigationGroup[] {
  const permissions = new Set(user.identity?.permissions ?? []);
  const role = user.activeRole?.code;
  const isFederationRole = role === "federation" || role === "super_admin";
  const items: NavigationItem[] = [
    { label: "Главная", icon: LayoutGrid, href: "/workspace" },
    { label: "Соревнования", icon: Trophy, href: "/workspace/competitions" },
    {
      label:
        permissions.has("rank_applications.review") ||
        permissions.has("judge_categories.review")
          ? "Заявки на проверку"
          : "Мои заявки",
      icon: ClipboardList,
      href: "/workspace/applications",
    },
    { label: "Важные сроки", icon: CalendarClock, href: "/workspace/deadlines" },
  ];

  if (role === "participant") {
    items.push(
      { label: "Мои тренеры", icon: UsersRound, href: "/workspace/coaches" },
      { label: "Мои результаты", icon: Award, href: "/workspace/results" },
      { label: "Спортивный разряд", icon: Medal, href: "/workspace/ranks" },
      { label: "Антидопинговые документы", icon: FileCheck2, href: "/workspace/documents" },
    );
  }
  if (role === "coach") {
    items.push(
      { label: "Мои спортсмены", icon: UsersRound, href: "/workspace/coaches" },
      { label: "Результаты спортсменов", icon: Award, href: "/workspace/results" },
    );
  }
  if (permissions.has("teams.manage") && !isFederationRole) {
    items.push({ label: "Сборные команды", icon: UsersRound, href: "/workspace/teams" });
  }
  if (role === "judge") {
    items.push({
      label: "Мои соревнования",
      icon: Gavel,
      href: "/workspace/judge-competitions",
    });
  }
  const administration: NavigationItem[] = [];
  if (permissions.has("registry.read")) {
    administration.push({
      label: "Реестр",
      icon: Building2,
      href: "/workspace/registry",
    });
  }
  if (permissions.has("calendar_plans.read")) {
    administration.push({
      label: "Календарные планы",
      icon: CalendarRange,
      href: "/workspace/calendar-plans",
    });
  }
  if (permissions.has("files.read")) {
    administration.push({ label: "Документы пользователей", icon: FileText, href: "/workspace/files" });
  }
  if (permissions.has("rating.read")) {
    administration.push({ label: "Рейтинг спортсменов", icon: Star, href: "/workspace/rating" });
  }
  if (permissions.has("teams.manage") && isFederationRole) {
    administration.push({ label: "Сборные команды", icon: UsersRound, href: "/workspace/teams" });
  }
  if (permissions.has("rank_applications.review")) {
    administration.push({ label: "Разряды и нормативы", icon: Medal, href: "/workspace/administration/rank-standards" });
  }
  const canAdminister =
    [
      "roles.read",
      "roles.create",
      "roles.assign",
      "permissions.read",
      "municipalities.create",
      "anti_doping.review",
      "rank_applications.review",
    ].some((permission) => permissions.has(permission));
  if (canAdminister) {
    if (permissions.has("roles.read") || permissions.has("roles.create")) {
      administration.push({ label: "Роли и права", icon: ShieldCheck, href: "/workspace/administration/roles" });
    }
    if (permissions.has("municipalities.create") || permissions.has("rank_applications.review")) {
      administration.push({ label: "Системные справочники", icon: ListTree, href: "/workspace/administration/directories" });
    }
    if (permissions.has("anti_doping.review")) {
      administration.push({ label: "Проверка сертификатов", icon: FileCheck2, href: "/workspace/administration/reviews" });
    }
    if (permissions.has("roles.read")) {
      administration.push({ label: "Журнал аудита", icon: ClipboardList, href: "/workspace/administration/audit" });
    }
  }
  return [
    { label: "Главное", items },
    ...(administration.length ? [{ label: "Управление федерацией", items: administration }] : []),
  ];
}

function NavigationSkeleton() {
  return (
    <SidebarMenu className="gap-1">
      {Array.from({ length: 4 }).map((_, index) => (
        <SidebarMenuItem key={index}>
          <div className="flex mb-1 px-2">
            <Skeleton
              className={cn(
                "h-9 rounded-md transition-opacity duration-150",
                "group-data-[collapsible=icon]:hidden w-full",
              )}
            />
          </div>
        </SidebarMenuItem>
      ))}
    </SidebarMenu>
  );
}

function WorkspaceNavigation({ user }: { user: WorkspaceUser }) {
  const { isMobile, openMobile, setOpenMobile } = useSidebar();
  const pathname = usePathname();
  const activeItemRef = useRef<HTMLAnchorElement | null>(null);

  const navigation = workspaceNavigation(user);

  useEffect(() => {
    if (!isMobile || !openMobile || !activeItemRef.current) return;
    const frame = requestAnimationFrame(() => {
      activeItemRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
    });
    return () => cancelAnimationFrame(frame);
  }, [isMobile, openMobile, pathname]);

  const isNavigationLoading = user.profileLoading;

  if (isNavigationLoading) {
    return <NavigationSkeleton />;
  }

  return (
    <div className="grid gap-4">
      {navigation.map((group) => (
        <div key={group.label} className="grid gap-1">
          <p className="px-2 text-xs font-medium text-muted-foreground group-data-[collapsible=icon]:hidden">{group.label}</p>
          <SidebarMenu className="gap-1">
            {group.items.map(({ label, icon: Icon, href }) => {
        const isActive = pathname === href;

        const content = (
          <>
            <Icon />

            <span className="text-sm leading-5 font-normal transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0">
              {label}
            </span>

            {isActive ? (
              <ChevronRight className="ml-auto transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0" />
            ) : null}
          </>
        );

              return (
          <SidebarMenuItem key={href}>
            <SidebarMenuButton
              asChild
              isActive={isActive}
              tooltip={label}
              className="h-10 data-[active=true]:bg-primary data-[active=true]:text-primary-foreground data-[active=true]:hover:bg-primary/90"
              onClick={() => {
                if (isMobile) {
                  setOpenMobile(false);
                }
              }}
            >
              <Link ref={isActive ? activeItemRef : undefined} href={href}>{content}</Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
              );
            })}
          </SidebarMenu>
        </div>
      ))}
    </div>
  );
}

function RoleDropdownItems({ user }: { user: WorkspaceUser }) {
  if (user.profileLoading) {
    return (
      <DropdownMenuItem disabled>
        <LoaderCircle className="animate-spin" />
        Загружаем роли…
      </DropdownMenuItem>
    );
  }

  if (!user.profile?.roles.length) {
    return (
      <DropdownMenuItem onSelect={() => user.retryProfile()}>
        {user.profileError ? "Профиль недоступен · повторить" : "Роли не назначены"}
      </DropdownMenuItem>
    );
  }

  return <>
    {!user.rolesCanBeReordered && user.profile.roles.length > 1 ? (
      <DropdownMenuLabel className="max-w-64 whitespace-normal text-xs font-normal text-muted-foreground">
        Для смены роли с административными правами обратитесь к администратору.
      </DropdownMenuLabel>
    ) : null}
    {user.profile.roles.map((role) => {
    const isActive = role.id === user.identity?.roleId;
    const isPending = role.id === user.pendingRoleId;

    return (
      <DropdownMenuItem
        key={role.id}
        disabled={
          user.pendingRoleId !== null ||
          (!isActive && !user.rolesCanBeReordered)
        }
        onSelect={() => void user.switchRole(role)}
      >
        {isPending ? (
          <LoaderCircle className="animate-spin" />
        ) : (
          <RoleIcon role={role} />
        )}
        <span className="flex-1">{role.name}</span>
        {isActive ? <Check className="text-primary!" /> : null}
      </DropdownMenuItem>
    );
    })}
  </>;
}

function DesktopUserMenu({ user }: { user: WorkspaceUser }) {
  const pathname = usePathname();
  const { theme, toggleTheme } = useTheme();
  const themeOrigin = useRef({ x: 0, y: 0 });

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          asChild
          isActive={pathname === "/workspace/notifications"}
          tooltip="Уведомления"
          className="h-10 data-[active=true]:bg-primary data-[active=true]:text-primary-foreground data-[active=true]:hover:bg-primary/90"
        >
          <Link href="/workspace/notifications">
            <Bell />
            <span className="transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0">
              Уведомления
            </span>
            {pathname === "/workspace/notifications" ? (
              <ChevronRight className="ml-auto transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0" />
            ) : null}
          </Link>
        </SidebarMenuButton>
      </SidebarMenuItem>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              tooltip="Профиль"
              className="p-0 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <ProfileAvatar name={user.displayName} />
              <div className="grid flex-1 text-left text-sm leading-tight transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0">
                <span className="truncate font-medium">{user.displayName}</span>
                <span className="truncate text-xs text-muted-foreground">
                  {user.profileLoading && "Загрузка…"}
                </span>
              </div>
              <ChevronsUpDown className="ml-auto size-4 transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 shadow-none"
            side="right"
            align="end"
            sideOffset={4}
          >
            <DropdownMenuItem asChild className="p-0 font-normal">
              <Link
                href="/workspace/profile"
                className="flex items-center gap-2 px-1 py-1.5 text-left text-sm"
              >
                <ProfileAvatar name={user.displayName} />
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">
                    {user.displayName}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">
                    {user.profileLoading && "Загрузка…"}
                  </span>
                </div>
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
                Доступные роли
              </DropdownMenuLabel>
              <RoleDropdownItems user={user} />
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onPointerDown={(event) => { themeOrigin.current = { x: event.clientX, y: event.clientY }; }} onSelect={() => toggleTheme(themeOrigin.current)}>
              {theme === "dark" ? <Sun /> : <Moon />}
              {theme === "dark" ? "Светлая тема" : "Тёмная тема"}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              variant="destructive"
              onSelect={() => void user.signOut()}
            >
              <LogOut />
              Выйти
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

function MobileSidebarFooter({ user }: { user: WorkspaceUser }) {
  const { setOpenMobile } = useSidebar();
  const { theme, toggleTheme } = useTheme();
  const themeOrigin = useRef({ x: 0, y: 0 });
  const pathname = usePathname();
  const mobileMenuButtonClass = "h-10 text-sm font-normal data-[active=true]:font-medium";

  return (
    <div className="flex flex-col gap-3">
      <SidebarMenu className="gap-1">
        <SidebarMenuItem>
          <SidebarMenuButton
            asChild
            isActive={pathname === "/workspace/notifications"}
            className={`${mobileMenuButtonClass} data-[active=true]:bg-primary data-[active=true]:text-primary-foreground data-[active=true]:hover:bg-primary/90`}
          >
            <Link href="/workspace/notifications" onClick={() => setOpenMobile(false)}>
              <Bell />
              <span className="text-sm leading-5 font-normal">Уведомления</span>
            </Link>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>

      <div>
        <p className="mb-1 px-2 text-xs text-muted-foreground">Выбор роли</p>
        <SidebarMenu className="gap-1">
          {user.profileLoading ? (
            <SidebarMenuItem>
              <SidebarMenuButton disabled className={mobileMenuButtonClass}>
                <LoaderCircle className="animate-spin" />
                <span className="text-sm leading-5 font-normal">Загружаем роли…</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ) : user.profile?.roles.length ? (
            user.profile.roles.map((role) => {
              const isActive = role.id === user.identity?.roleId;
              const isPending = role.id === user.pendingRoleId;

              return (
                <SidebarMenuItem key={role.id}>
                  <SidebarMenuButton
                    isActive={isActive}
                    disabled={
                      user.pendingRoleId !== null ||
                      (!isActive && !user.rolesCanBeReordered)
                    }
                    className={`${mobileMenuButtonClass} data-[active=true]:bg-primary data-[active=true]:text-primary-foreground data-[active=true]:hover:bg-primary/90`}
                    onClick={() => void user.switchRole(role)}
                  >
                    {isPending ? (
                      <LoaderCircle className="animate-spin" />
                    ) : (
                      <RoleIcon role={role} />
                    )}
                    <span className="flex-1 text-sm leading-5 font-normal">{role.name}</span>
                    {isActive ? <Check /> : null}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              );
            })
          ) : user.profileError ? (
            <SidebarMenuItem><SidebarMenuButton className={mobileMenuButtonClass} onClick={() => user.retryProfile()}><ShieldCheck /><span className="text-sm leading-5 font-normal">Повторить загрузку профиля</span></SidebarMenuButton></SidebarMenuItem>
          ) : (
            <SidebarMenuItem>
              <SidebarMenuButton disabled className={mobileMenuButtonClass}>
                <ShieldCheck />
                <span className="text-sm leading-5 font-normal">Роли пока не назначены</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          )}
        </SidebarMenu>
      </div>

      <SidebarSeparator className="mx-0" />
      <Button
        variant="ghost"
        className="h-10 w-full justify-start px-2 text-sm font-normal"
        onClick={() => toggleTheme(themeOrigin.current)}
        onPointerDown={(event) => { themeOrigin.current = { x: event.clientX, y: event.clientY }; }}
      >
        {theme === "dark" ? <Sun /> : <Moon />}
        <span className="text-sm leading-5 font-normal">{theme === "dark" ? "Светлая тема" : "Тёмная тема"}</span>
      </Button>
      <Button
        variant="ghost"
        className="h-10 w-full justify-start px-2 text-sm font-normal text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => {
          setOpenMobile(false);
          void user.signOut();
        }}
      >
        <LogOut />
        <span className="text-sm leading-5 font-normal">Выйти</span>
      </Button>
    </div>
  );
}

function MobileHeader({ user }: { user: WorkspaceUser }) {
  const { theme } = useTheme();

  return (
    <header className="grid h-16 shrink-0 grid-cols-[40px_1fr_auto] items-center gap-2 px-4 md:hidden">
      <SidebarTrigger
        aria-label="Открыть меню"
        className="size-9 text-foreground hover:bg-muted"
      >
        <Menu className="size-6" />
      </SidebarTrigger>
      <BrandLogo
        href="/workspace"
        theme={theme === "dark" ? "light" : "dark"}
        className="mx-auto h-7 gap-2.5 [&_img:first-child]:size-7 [&_img:last-child]:h-auto [&_img:last-child]:w-10"
      />
      <div className="flex items-center gap-1">
        <Button
          asChild
          variant="ghost"
          size="icon"
          className="size-9 p-0 hover:bg-muted"
        >
          <Link href="/workspace/profile" aria-label="Открыть профиль">
            <ProfileAvatar
              name={user.displayName}
              className="size-9 rounded-full border border-border/80"
            />
          </Link>
        </Button>
      </div>
    </header>
  );
}

export function WorkspaceShell({
  children,
}: {
  children?: ReactNode;
}) {
  const user = useWorkspaceUser();
  const pathname = usePathname();
  const title = pageTitle(pathname);
  const { theme } = useTheme();

  return (
    <main className="relative h-svh min-w-0 overflow-hidden bg-workspace">
      <div className="absolute inset-0">
        {theme === "dark" ? <div className="opacity-20"> <RibbonGlowDark /></div> : <RibbonGlow />}
      </div>
      <div className="relative z-10 h-full">
        <SidebarProvider
          defaultOpen
          style={
            {
              "--sidebar-width": "18rem",
              "--sidebar-width-icon": "65px",
            } as CSSProperties
          }
        >
          <Sidebar
            collapsible="icon"
            className="border-0 max-md:shadow-none [&_[data-sidebar=sidebar]]:border-r [&_[data-sidebar=sidebar]]:border-border [&_[data-sidebar=sidebar]]:bg-sidebar"
          >
            <SidebarHeader className="hidden gap-2 px-5 py-2 md:flex">
              <div className="flex h-10 items-center">
                <BrandLogo
                  href="/workspace"
                  theme={theme === "dark" ? "light" : "dark"}
                  className="h-6 w-auto group-data-[collapsible=icon]:gap-0 group-data-[collapsible=icon]:[&_img:last-child]:hidden"
                />
              </div>
            </SidebarHeader>

            <SidebarContent className="px-4 py-5 md:py-1">
              <WorkspaceNavigation user={user} />
            </SidebarContent>

            <SidebarFooter className="gap-2 px-4 py-4 md:py-2">
              <div className="md:hidden">
                <MobileSidebarFooter user={user} />
              </div>
              <div className="hidden md:block">
                <DesktopUserMenu user={user} />
              </div>
            </SidebarFooter>
          </Sidebar>

          <SidebarInset className="h-svh min-w-0 bg-background/40 backdrop-blur-[20px] dark:backdrop-blur-[36px]">
            <MobileHeader user={user} />
            <header className="hidden h-14 shrink-0 items-center gap-3 px-3 md:flex">
              <SidebarTrigger aria-label="Свернуть боковое меню" />
              <h1 className="text-base font-semibold leading-6 tracking-[-0.01em]">{title}</h1>
            </header>
            <div className="px-4 pb-2 text-center md:hidden">
              <h1 className="text-base leading-6 font-semibold tracking-[-0.01em]">
                {title}
              </h1>
            </div>
            <div data-workspace-toolbar-host className="shrink-0 px-4 md:px-6" />
            {children ? (
              <div className="relative min-h-0 flex-1">
                <ScrollMask className="h-full p-4 pb-7 pt-4 md:px-6 md:pb-9 md:pt-5" fade={32}>
                  <div className="mx-auto w-full max-w-6xl">
                    {children}
                  </div>
                </ScrollMask>
              </div>
            ) : null}
          </SidebarInset>
        </SidebarProvider>
      </div>
    </main>
  );
}
