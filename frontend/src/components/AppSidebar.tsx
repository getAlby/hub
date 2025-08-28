import {
  BoxIcon,
  ChevronsUpDown,
  CircleHelp,
  HomeIcon,
  LogOut,
  LucideIcon,
  Plug2Icon,
  PlugZapIcon,
  Settings,
  Sparkles,
  SquareStack,
  StarIcon,
  WalletIcon,
} from "lucide-react";
import React from "react";

import { Link, NavLink, useNavigate } from "react-router-dom";
import albyHub from "src/assets/suggested-apps/alby-hub.png";
import ExternalLink from "src/components/ExternalLink";
import { AlbyIcon } from "src/components/icons/Alby";
import { AlbyHubLogo } from "src/components/icons/AlbyHubLogo";
import SidebarHint from "src/components/SidebarHint";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "src/components/ui/sidebar";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import UserAvatar from "src/components/UserAvatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useHealthCheck } from "src/hooks/useHealthCheck";
import { useInfo } from "src/hooks/useInfo";
import { deleteAuthToken } from "src/lib/auth";
import { isHttpMode } from "src/utils/isHttpMode";

export function AppSidebar() {
  const { data: albyMe } = useAlbyMe();

  const { data: info, mutate: refetchInfo } = useInfo();
  const { isMobile, setOpenMobile } = useSidebar();
  const { hasChannelManagement } = useInfo();
  const navigate = useNavigate();

  const _isHttpMode = isHttpMode();

  const logout = React.useCallback(async () => {
    deleteAuthToken();
    await refetchInfo();

    if (_isHttpMode) {
      window.location.href = "/logout";
    } else {
      navigate("/", { replace: true });
    }
  }, [_isHttpMode, navigate, refetchInfo]);

  const data = {
    navMain: [
      {
        title: "Home",
        url: "/home",
        icon: HomeIcon,
      },
      {
        title: "Wallet",
        url: "/wallet",
        icon: WalletIcon,
      },
      {
        title: "Sub-wallets",
        url: "/sub-wallets",
        icon: SquareStack,
      },
      {
        title: "Connections",
        url: "/apps",
        icon: Plug2Icon,
      },
    ],
    navSecondary: [
      ...(hasChannelManagement
        ? [
            {
              title: "Node",
              url: "/channels",
              icon: BoxIcon,
            },
          ]
        : []),
      {
        title: "Settings",
        url: "/settings",
        icon: Settings,
      },
      {
        title: "Review & Earn",
        url: "/review-earn",
        icon: StarIcon,
      },
    ],
  };

  return (
    <Sidebar
      className="top-(--header-height) h-[calc(100svh-var(--header-height))]!"
      collapsible="offcanvas"
    >
      <SidebarHeader>
        <div className="p-2 flex flex-row items-center justify-between">
          <Link to="/home" onClick={() => setOpenMobile(false)}>
            <AlbyHubLogo className="text-sidebar-foreground h-12" />
          </Link>
          <div className="flex gap-3 items-center">
            <HealthIndicator />
          </div>
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {data.navMain.map((item) => (
                <NavLink key={item.title} to={item.url} end>
                  {({ isActive }) => (
                    <SidebarMenuItem
                      onClick={() => {
                        setOpenMobile(false);
                      }}
                    >
                      <SidebarMenuButton isActive={isActive}>
                        <item.icon />
                        <span>{item.title}</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )}
                </NavLink>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <div className="mt-auto">
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarHint />
            </SidebarGroupContent>
          </SidebarGroup>
          <NavSecondary items={data.navSecondary} />
        </div>
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger className="w-full">
                <SidebarMenuButton size="lg">
                  {info?.albyAccountConnected ? (
                    <>
                      <UserAvatar />
                      <div className="grid flex-1 text-left text-sm leading-tight">
                        <span className="truncate font-semibold">
                          {albyMe?.name || albyMe?.email}
                        </span>
                        <div className="truncate text-xs">
                          {albyMe?.lightning_address}
                        </div>
                      </div>
                    </>
                  ) : (
                    <>
                      <img
                        src={albyHub}
                        alt="logo"
                        className="size-8 rounded-lg "
                      />
                      <div className="font-semibold text-left text-sm leading-tight">
                        My Alby Hub
                      </div>
                    </>
                  )}
                  <ChevronsUpDown className="ml-auto size-4" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="min-w-56"
                side={isMobile ? "bottom" : "right"}
                align="end"
                sideOffset={4}
              >
                {info?.albyAccountConnected && (
                  <>
                    <DropdownMenuLabel className="p-0 font-normal">
                      <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                        <UserAvatar />
                        <div className="grid flex-1 text-left text-sm leading-tight">
                          <span className="truncate font-semibold">
                            {albyMe?.name || albyMe?.email}
                          </span>
                          <span className="truncate text-xs">
                            {albyMe?.lightning_address}
                          </span>
                        </div>
                      </div>
                    </DropdownMenuLabel>
                    <DropdownMenuSeparator />
                  </>
                )}
                {!info?.albyAccountConnected ? (
                  <DropdownMenuItem asChild>
                    <Link
                      to="/alby/account"
                      className="w-full flex flex-row items-center gap-2"
                    >
                      <PlugZapIcon />
                      Connect Alby Account
                    </Link>
                  </DropdownMenuItem>
                ) : (
                  <DropdownMenuItem>
                    <ExternalLink
                      to="https://getalby.com/user/edit"
                      className="flex items-center flex-1"
                    >
                      <AlbyIcon className="size-4 mr-2" />
                      Alby Account Settings
                    </ExternalLink>
                  </DropdownMenuItem>
                )}
                {!albyMe?.subscription.plan_code && (
                  <>
                    <UpgradeDialog>
                      <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                        <Sparkles />
                        Upgrade to Pro
                      </DropdownMenuItem>
                    </UpgradeDialog>
                  </>
                )}
                {_isHttpMode && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={logout}>
                      <LogOut className="size-4 mr-2" />
                      Log out
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}

export function NavSecondary({
  items,
  ...props
}: {
  items: {
    title: string;
    url: string;
    icon: LucideIcon;
  }[];
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const { data: albyMe } = useAlbyMe();
  const { data: info } = useInfo();
  const { setOpenMobile } = useSidebar();

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <NavLink key={item.title} to={item.url} end>
              {({ isActive }) => (
                <SidebarMenuItem
                  onClick={() => {
                    setOpenMobile(false);
                  }}
                >
                  <SidebarMenuButton isActive={isActive}>
                    <item.icon />
                    <span>{item.title}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}
            </NavLink>
          ))}
          <ExternalLink to="https://support.getalby.com">
            <SidebarMenuItem>
              <SidebarMenuButton>
                <CircleHelp className="h-4 w-4" />
                Help
              </SidebarMenuButton>
            </SidebarMenuItem>
          </ExternalLink>
          {!albyMe?.subscription.plan_code && info?.albyAccountConnected && (
            <ExternalLink to="https://getalby.com/subscription/pro">
              <SidebarMenuItem>
                <SidebarMenuButton>
                  <Sparkles className="h-4 w-4" />
                  Alby Pro
                </SidebarMenuButton>
              </SidebarMenuItem>
            </ExternalLink>
          )}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}

function HealthIndicator() {
  const { data: health } = useHealthCheck();
  const { setOpenMobile } = useSidebar();

  if (!health) {
    return null;
  }

  const ok = !health.alarms?.length;
  if (ok) {
    return null;
  }

  return (
    <Link to="/channels" onClick={() => setOpenMobile(false)}>
      <div className="w-2 h-2 rounded-full bg-destructive" />
    </Link>
  );
}
