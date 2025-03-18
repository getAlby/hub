import {
  BoxIcon,
  ChevronsUpDown,
  CircleHelp,
  Cloud,
  Home,
  HomeIcon,
  LayoutGridIcon,
  LogOut,
  LucideIcon,
  Plug2Icon,
  PlugZapIcon,
  Settings,
  Sparkles,
  UserCog,
  WalletIcon,
} from "lucide-react";
import React from "react";

import { Link, NavLink, useNavigate } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import { AlbyHubLogo } from "src/components/icons/AlbyHubLogo";
import SidebarHint from "src/components/SidebarHint";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
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
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import UserAvatar from "src/components/UserAvatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useHealthCheck } from "src/hooks/useHealthCheck";
import { useInfo } from "src/hooks/useInfo";
import { deleteAuthToken } from "src/lib/auth";
import { cn } from "src/lib/utils";
import { HealthAlarm } from "src/types";

import { isHttpMode } from "src/utils/isHttpMode";

// Menu items.
const items = [
  {
    title: "Home",
    url: "#",
    icon: Home,
  },
];

export function AppSidebar() {
  const { data: albyMe } = useAlbyMe();

  const { data: info, mutate: refetchInfo } = useInfo();
  const { isMobile } = useSidebar();
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
    user: {
      name: "shadcn",
      email: "m@example.com",
      avatar: "/avatars/shadcn.jpg",
    },
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
        title: "App Store",
        url: "/appstore",
        icon: LayoutGridIcon,
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
    ],
  };

  return (
    <Sidebar variant="sidebar">
      <SidebarHeader className="p-5 flex flex-row items-center justify-between">
        <Link to="/">
          <AlbyHubLogo className="text-sidebar-foreground h-12" />
        </Link>
        <HealthIndicator />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {data.navMain.map((item) => (
                <NavLink key={item.title} to={item.url} end>
                  {({ isActive }) => (
                    <SidebarMenuItem>
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
        <SidebarGroup className="mt-auto">
          <SidebarHint />
          <NavSecondary items={data.navSecondary} />
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  {info?.albyAccountConnected && (
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
                  )}
                  <ChevronsUpDown className="ml-auto size-4" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
                side={isMobile ? "bottom" : "right"}
                align="end"
                sideOffset={4}
              >
                {info?.albyAccountConnected && (
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
                )}
                {!info?.albyAccountConnected && (
                  <DropdownMenuItem>
                    <Link
                      to="/alby/account"
                      className="w-full flex flex-row items-center gap-2"
                    >
                      <PlugZapIcon className="w-4 h-4" />
                      <p>Connect Alby Account</p>
                    </Link>
                  </DropdownMenuItem>
                )}
                {!albyMe?.subscription.buzz && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuGroup>
                      <DropdownMenuItem>
                        {/* TODO: add upgradedialog */}
                        <Sparkles className="w-4 h-4 mr-2" />
                        Upgrade to Pro
                      </DropdownMenuItem>
                    </DropdownMenuGroup>
                  </>
                )}
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuItem>
                    <ExternalLink
                      to="https://getalby.com/user/edit"
                      className="flex items-center"
                    >
                      <UserCog className="w-4 h-4 mr-2" />
                      Manage Account
                    </ExternalLink>
                  </DropdownMenuItem>
                </DropdownMenuGroup>
                {_isHttpMode && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={logout}>
                      <LogOut className="w-4 h-4 mr-2" />
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

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <NavLink key={item.title} to={item.url} end>
              {({ isActive }) => (
                <SidebarMenuItem>
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
          {/* Does his still make sense? Can we limit that to desktop users? */}
          {!albyMe?.hub.name && info?.albyAccountConnected && (
            <ExternalLink to="https://getalby.com/subscription/new">
              <SidebarMenuItem>
                <SidebarMenuButton>
                  <Cloud className="h-4 w-4" />
                  Alby Cloud
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
  if (!health) {
    return null;
  }

  const ok = !health.alarms?.length;

  function getAlarmTitle(alarm: HealthAlarm) {
    // TODO: could show extra data from alarm.rawDetails
    // for some alarm types
    switch (alarm.kind) {
      case "alby_service":
        return "One or more Alby Services are offline";
      case "channels_offline":
        return "One or more channels are offline";
      case "node_not_ready":
        return "Node is not ready";
      case "nostr_relay_offline":
        return "Could not connect to relay";
      default:
        return "Unknown error";
    }
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <div className="w-8 h-8 flex items-center justify-center">
            <span className="text-xs flex items-center text-muted-foreground">
              <div
                className={cn(
                  "w-2 h-2 rounded-full",
                  ok ? "bg-green-300" : "bg-destructive"
                )}
              />
            </span>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          {ok ? (
            <p>Alby Hub is running</p>
          ) : (
            <div>
              <p className="font-semibold">
                {health.alarms.length} issues were found
              </p>
              <ul className="mt-2 max-w-xs whitespace-pre-wrap list-disc list-inside">
                {health.alarms.map((alarm) => (
                  <li key={alarm.kind}>{getAlarmTitle(alarm)}</li>
                ))}
              </ul>
            </div>
          )}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
