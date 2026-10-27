"use client";

import type { CSSProperties } from "react";
import {
  Building2,
  ChevronRight,
  ChevronsUpDown,
  LayoutGrid,
  LogOut,
  Settings,
  Smile,
  UserRound,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/entities/session";
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar";
import { BrandLogo } from "@/shared/ui/brand-logo";
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

const navigation = [
  "Обзор",
  "Соревнования",
  "Мои заявки",
  "Результаты",
  "Рейтинг",
  "Мой разряд",
];
const roles = [
  [UserRound, "Спортсмен"],
  [Smile, "Тренер"],
  [Smile, "Судья"],
  [Building2, "Федерация"],
] as const;

function UserMenu() {
  const { isMobile } = useSidebar();
  const { logout } = useAuth();
  const router = useRouter();
  const signOut = async () => {
    await logout();
    router.replace("/login");
  };
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              tooltip="Профиль"
              className="p-0 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar className="size-8 rounded-lg">
                <AvatarImage
                  src="/assets/workspace-avatar.png"
                  alt="Павел"
                  className="rounded-lg object-cover grayscale"
                />
                <AvatarFallback className="rounded-lg bg-[#e8e5ff] text-[#402fff]">
                  <UserRound className="size-4" />
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0">
                <span className="truncate font-medium">Павел</span>
                <span className="truncate text-xs text-muted-foreground">
                  Спортсмен
                </span>
              </div>
              <ChevronsUpDown className="ml-auto size-4 transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="size-8 rounded-lg">
                  <AvatarImage
                    src="/assets/workspace-avatar.png"
                    alt="Павел"
                    className="rounded-lg object-cover grayscale"
                  />
                  <AvatarFallback className="rounded-lg bg-[#e8e5ff] text-[#402fff]">
                    <UserRound className="size-4" />
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">Павел</span>
                  <span className="truncate text-xs text-muted-foreground">
                    Спортсмен
                  </span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
                Переключить роль
              </DropdownMenuLabel>
              {roles.map(([Icon, title]) => (
                <DropdownMenuItem key={title}>
                  <Icon />
                  {title}
                </DropdownMenuItem>
              ))}
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem variant="destructive" onSelect={signOut}>
              <LogOut />
              Выйти
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

export function WorkspaceShell() {
  return (
    <main className="relative h-svh min-w-[1200px] overflow-hidden bg-[#f2efff]">
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
            className="border-0 [&_[data-sidebar=sidebar]]:bg-white"
          >
            <SidebarHeader className="gap-2 px-5 py-2 ">
              <div className="flex h-10 items-center">
                <BrandLogo className="h-6 w-auto group-data-[collapsible=icon]:gap-0 group-data-[collapsible=icon]:[&_img:last-child]:hidden" />
              </div>
            </SidebarHeader>
            <SidebarContent className="px-4">
              <SidebarMenu className="gap-1">
                {navigation.map((item, index) => (
                  <SidebarMenuItem key={item}>
                    <SidebarMenuButton
                      isActive={index === 0}
                      tooltip={item}
                      className="data-[active=true]:bg-[#402fff] data-[active=true]:text-white data-[active=true]:hover:bg-[#402fff]/90"
                    >
                      <LayoutGrid />{" "}
                      <span className="transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0">
                        {item}
                      </span>
                      {index === 0 && (
                        <ChevronRight className="ml-auto transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0" />
                      )}
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarContent>
            <SidebarFooter className="gap-2 px-4 py-2">
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton tooltip="Настройки">
                    <Settings />
                    <span className="transition-opacity duration-150 group-data-[collapsible=icon]:opacity-0">
                      Настройки
                    </span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
              <SidebarSeparator className="mx-0 group-data-[collapsible=icon]:hidden" />
              <UserMenu />
            </SidebarFooter>
          </Sidebar>
          <SidebarInset className="h-svh min-w-0 bg-white/40 backdrop-blur-[20px]">
            <header className="flex h-16 shrink-0 items-center gap-3 px-3">
              <SidebarTrigger aria-label="Свернуть боковое меню" />
              <h1 className="text-sm font-medium">Обзор</h1>
            </header>
          </SidebarInset>
        </SidebarProvider>
      </div>
    </main>
  );
}
