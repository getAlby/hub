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
  Settings,
  Sparkles,
  UserCog,
  WalletIcon,
} from "lucide-react";

import { Link, NavLink } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import { AlbyHubLogo } from "src/components/icons/AlbyHubLogo";
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
import UserAvatar from "src/components/UserAvatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

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
      <SidebarHeader className="p-5">
        <Link to="/">
          <AlbyHubLogo className="text-sidebar-foreground h-12" />
        </Link>
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
        <NavSecondary items={data.navSecondary} className="mt-auto" />
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
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <LogOut className="w-4 h-4 mr-2" />
                  Log out
                </DropdownMenuItem>
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
                Support
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
