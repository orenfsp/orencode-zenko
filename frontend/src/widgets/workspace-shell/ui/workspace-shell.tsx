"use client";

import {
  useEffect,
  useMemo,
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
  ShieldCheck,
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
import { Skeleton } from "@/shared/ui/skeleton";
import { cn } from "@/shared/lib/utils";

interface NavigationItem {
  label: string;
  icon: LucideIcon;
  href: string;
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
      <AvatarFallback className="rounded-[inherit] bg-[#e8e5ff] text-[#402fff]">
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
  const [pendingRoleId, setPendingRoleId] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;

    queueMicrotask(() => {
      if (cancelled) return;
      setProfile(null);
      setProfileLoading(true);
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
      } catch (error) {
        if (cancelled) return;

        console.error(
          "[WorkspaceShell] Не удалось загрузить профиль. Повтор через 3 секунды:",
          error,
        );

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
  }, [identity]);

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
    await logout();
    router.replace("/login");
  }

  return {
    identity,
    profile,
    profileLoading,
    pendingRoleId,
    activeRole,
    displayName: profileName(profile),
    rolesCanBeReordered,
    switchRole,
    signOut,
  };
}

type WorkspaceUser = ReturnType<typeof useWorkspaceUser>;

function workspaceNavigation(user: WorkspaceUser): NavigationItem[] {
  const permissions = new Set(user.identity?.permissions ?? []);
  const role = user.activeRole?.code;
  const isFederationRole = role === "federation" || role === "super_admin";
  const items: NavigationItem[] = [
    { label: "Обзор", icon: LayoutGrid, href: "/workspace" },
    { label: "Соревнования", icon: Trophy, href: "/workspace/competitions" },
    {
      label:
        isFederationRole ||
        permissions.has("rank_applications.review") ||
        permissions.has("competition_applications.review") ||
        permissions.has("judge_categories.review")
          ? "Заявки"
          : "Мои заявки",
      icon: ClipboardList,
      href: "/workspace/applications",
    },
    { label: "Сроки", icon: CalendarClock, href: "/workspace/deadlines" },
  ];

  if (role === "participant") {
    items.push(
      { label: "Мои тренеры", icon: UsersRound, href: "/workspace/coaches" },
      { label: "Результаты", icon: Award, href: "/workspace/results" },
      { label: "Мой разряд", icon: Medal, href: "/workspace/ranks" },
      { label: "Антидопинг", icon: FileCheck2, href: "/workspace/documents" },
    );
  }
  if (role === "coach") {
    items.push(
      { label: "Мои спортсмены", icon: UsersRound, href: "/workspace/coaches" },
      { label: "Результаты", icon: Award, href: "/workspace/results" },
    );
  }
  if (role === "judge") {
    items.push({
      label: "Мои соревнования",
      icon: Gavel,
      href: "/workspace/judge-competitions",
    });
  }
  if (isFederationRole || permissions.has("registry.read")) {
    items.push({
      label: "Реестр",
      icon: Building2,
      href: "/workspace/registry",
    });
  }
  if (isFederationRole || permissions.has("calendar_plans.read")) {
    items.push({
      label: "Календарные планы",
      icon: CalendarRange,
      href: "/workspace/calendar-plans",
    });
  }
  if (isFederationRole || permissions.has("competition_events.read")) {
    items.push({ label: "Зачёты", icon: ListTree, href: "/workspace/events" });
  }
  if (isFederationRole || permissions.has("files.read")) {
    items.push({ label: "Файлы", icon: FileText, href: "/workspace/files" });
  }
  if (
    isFederationRole ||
    [
      "roles.read",
      "roles.create",
      "roles.assign",
      "permissions.read",
      "municipalities.create",
      "anti_doping.review",
      "rank_applications.review",
    ].some((permission) => permissions.has(permission))
  ) {
    items.push({
      label: "Администрирование",
      icon: ShieldCheck,
      href: "/workspace/administration",
    });
  }
  return items;
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
  const { isMobile, setOpenMobile } = useSidebar();
  const pathname = usePathname();

  const navigation = workspaceNavigation(user);

  const isNavigationLoading =
    user.profileLoading || !user.profile || !user.activeRole;

  if (isNavigationLoading) {
    return <NavigationSkeleton />;
  }

  return (
    <SidebarMenu className="gap-1">
      {navigation.map(({ label, icon: Icon, href }) => {
        const isActive = pathname === href;

        const content = (
          <>
            <Icon />

            <span className="transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0">
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
              className="h-10 data-[active=true]:bg-[#402fff] data-[active=true]:text-white data-[active=true]:hover:bg-[#402fff]/90"
              onClick={() => {
                if (isMobile) {
                  setOpenMobile(false);
                }
              }}
            >
              <Link href={href}>{content}</Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        );
      })}
    </SidebarMenu>
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
      <DropdownMenuItem disabled>
        <LoaderCircle className="animate-spin" />
        Загружаем роли…
      </DropdownMenuItem>
    );
  }

  return user.profile.roles.map((role) => {
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
  });
}

function DesktopUserMenu({ user }: { user: WorkspaceUser }) {
  const pathname = usePathname();

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <SidebarMenuButton
          asChild
          isActive={pathname === "/workspace/notifications"}
          tooltip="Уведомления"
          className="h-10 data-[active=true]:bg-[#402fff] data-[active=true]:text-white data-[active=true]:hover:bg-[#402fff]/90"
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
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg shadow-none"
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
  const pathname = usePathname();

  return (
    <div className="flex flex-col gap-3">
      <SidebarMenu className="gap-1">
        <SidebarMenuItem>
          <SidebarMenuButton
            asChild
            isActive={pathname === "/workspace/notifications"}
            className="h-10 data-[active=true]:bg-[#f0eeff] data-[active=true]:text-[#402fff]"
          >
            <Link href="/workspace/notifications" onClick={() => setOpenMobile(false)}>
              <Bell />
              <span>Уведомления</span>
            </Link>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>

      <div>
        <p className="mb-1 px-2 text-xs text-muted-foreground">Выбор роли</p>
        <SidebarMenu className="gap-1">
          {user.profileLoading ? (
            <SidebarMenuItem>
              <SidebarMenuButton disabled>
                <LoaderCircle className="animate-spin" />
                <span>Загружаем роли…</span>
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
                    className="h-10 data-[active=true]:bg-[#f0eeff] data-[active=true]:text-[#402fff]"
                    onClick={() => void user.switchRole(role)}
                  >
                    {isPending ? (
                      <LoaderCircle className="animate-spin" />
                    ) : (
                      <RoleIcon role={role} />
                    )}
                    <span className="flex-1">{role.name}</span>
                    {isActive ? <Check /> : null}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              );
            })
          ) : (
            <SidebarMenuItem>
              <SidebarMenuButton disabled>
                <ShieldCheck />
                <span>Роли пока не назначены</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          )}
        </SidebarMenu>
      </div>

      <SidebarSeparator className="mx-0" />
      <Button
        variant="ghost"
        className="h-10 w-full justify-start px-2 text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => {
          setOpenMobile(false);
          void user.signOut();
        }}
      >
        <LogOut />
        Выйти
      </Button>
    </div>
  );
}

function MobileHeader({ user }: { user: WorkspaceUser }) {
  return (
    <header className="grid h-[72px] shrink-0 grid-cols-[40px_1fr_40px] items-center px-4 md:hidden">
      <SidebarTrigger
        aria-label="Открыть меню"
        className="size-10 rounded-full text-foreground hover:bg-white/70"
      >
        <Menu className="size-6" />
      </SidebarTrigger>
      <BrandLogo
        href="/workspace"
        className="mx-auto h-7 gap-2.5 [&_img:first-child]:size-7 [&_img:last-child]:h-auto [&_img:last-child]:w-10"
      />
      <Button
        asChild
        variant="ghost"
        size="icon"
        className="size-10 rounded-full p-0 hover:bg-white/70"
      >
        <Link href="/workspace/profile" aria-label="Открыть профиль">
          <ProfileAvatar
            name={user.displayName}
            className="size-9 rounded-full border border-white/80"
          />
        </Link>
      </Button>
    </header>
  );
}

export function WorkspaceShell({
  children,
  title = "Обзор",
}: {
  children?: ReactNode;
  title?: string;
}) {
  const user = useWorkspaceUser();

  return (
    <main className="relative h-svh min-w-0 overflow-hidden bg-[#f2efff] md:min-w-[1200px]">
      <div className="absolute inset-0">
        <RibbonGlow />
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
            className="border-0 max-md:shadow-none [&_[data-sidebar=sidebar]]:bg-white"
          >
            <SidebarHeader className="hidden gap-2 px-5 py-2 md:flex">
              <div className="flex h-10 items-center">
                <BrandLogo
                  href="/workspace"
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

          <SidebarInset className="h-svh min-w-0 bg-white/40 backdrop-blur-[20px]">
            <MobileHeader user={user} />
            <header className="hidden h-16 shrink-0 items-center gap-3 px-3 md:flex">
              <SidebarTrigger aria-label="Свернуть боковое меню" />
              <h1 className="text-sm font-medium">{title}</h1>
            </header>
            <div className="px-4 pb-5 text-center md:hidden">
              <h1 className="text-xl leading-7 font-medium tracking-[-0.01em]">
                {title}
              </h1>
            </div>
            {children ? (
              <section className="min-h-0 flex-1 overflow-y-auto p-4 pt-0 md:p-6 md:pt-0">
                {children}
              </section>
            ) : null}
          </SidebarInset>
        </SidebarProvider>
      </div>
    </main>
  );
}
