import {
  Cable,
  EllipsisVertical,
  ExternalLinkIcon,
  FlaskRound,
  Lock,
  Megaphone,
  Menu,
  MessageCircleQuestion,
  Settings,
  ShieldAlertIcon,
  ShieldCheckIcon,
  Store,
  Wallet,
} from "lucide-react";

import React from "react";
import {
  Link,
  NavLink,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom";
import SidebarHint from "src/components/SidebarHint";
import UserAvatar from "src/components/UserAvatar";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetTrigger } from "src/components/ui/sheet";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { useRemoveSuccessfulChannelOrder } from "src/hooks/useRemoveSuccessfulChannelOrder";
import { cn } from "src/lib/utils";
import { openLink } from "src/utils/openLink";
import { request } from "src/utils/request";
import ExternalLink from "../ExternalLink";

export default function AppLayout() {
  const { data: albyMe } = useAlbyMe();
  const { data: csrf } = useCSRF();
  const { data: info, mutate: refetchInfo } = useInfo();
  const { toast } = useToast();
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  useRemoveSuccessfulChannelOrder();

  React.useEffect(() => {
    setMobileMenuOpen(false);
  }, [location]);

  const logout = React.useCallback(async () => {
    if (!csrf) {
      throw new Error("csrf not loaded");
    }

    await request("/api/logout", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
    });

    await refetchInfo();
    navigate("/", { replace: true });
    toast({ title: "You are now logged out." });
  }, [csrf, navigate, refetchInfo, toast]);

  const isHttpMode = window.location.protocol.startsWith("http");

  function UserMenuContent() {
    return (
      <DropdownMenuContent align="end">
        <DropdownMenuGroup>
          <DropdownMenuItem>
            <ExternalLink
              to="https://getalby.com/settings"
              className="w-full flex flex-row items-center gap-2"
            >
              <ExternalLinkIcon className="w-4 h-4" />
              <p>Alby Account Settings</p>
            </ExternalLink>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        {isHttpMode && (
          <DropdownMenuItem
            onClick={logout}
            className="w-full flex flex-row items-center gap-2"
          >
            <Lock className="w-4 h-4" />
            <p>Lock Alby Hub</p>
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    );
  }

  function MainMenuContent() {
    return (
      <>
        <MenuItem to="/wallet">
          <Wallet className="h-4 w-4" />
          Wallet
        </MenuItem>
        <MenuItem to="/apps">
          <Cable className="h-4 w-4" />
          Connections
        </MenuItem>
        <MenuItem to="/appstore">
          <Store className="h-4 w-4" />
          App Store
        </MenuItem>
      </>
    );
  }

  function MainNavSecondary() {
    const { hasChannelManagement } = useInfo();
    return (
      <nav className="grid items-start p-2 text-sm font-medium lg:px-4">
        {/* <div className="px-3 py-2 mb-5">
          <ModeToggle />
        </div> */}
        {hasChannelManagement && (
          <MenuItem to="/channels">
            <FlaskRound className="h-4 w-4" />
            Liquidity
          </MenuItem>
        )}
        <MenuItem to="/settings">
          <Settings className="h-4 w-4" />
          Settings
        </MenuItem>
        <MenuItem
          to="/"
          onClick={(e) => {
            openLink(
              "https://feedback.getalby.com/-alby-hub-request-a-feature"
            );
            e.preventDefault();
          }}
        >
          <Megaphone className="h-4 w-4" />
          Feedback
        </MenuItem>
        <MenuItem
          to="/"
          onClick={(e) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const chatwoot = (window as any).$chatwoot;
            if (chatwoot) {
              chatwoot.toggle("open");
            } else {
              openLink("https://getalby.com/help");
            }

            e.preventDefault();
          }}
        >
          <MessageCircleQuestion className="h-4 w-4" />
          Live Support
        </MenuItem>
      </nav>
    );
  }

  const upToDate = info?.version && info.version === info.latestVersion;

  return (
    <>
      <div className="font-sans min-h-screen w-full flex flex-col">
        <div className="flex-1 h-full grid md:grid-cols-[220px_1fr] lg:grid-cols-[280px_1fr]">
          <div className="hidden border-r bg-muted/40 md:block">
            <div className="flex h-full max-h-screen flex-col gap-2 sticky top-0">
              <div className="flex-1">
                <nav className="grid items-start px-2 py-2 text-sm font-medium lg:px-4">
                  <div className="p-3 flex justify-between items-center">
                    <Link to="/" className="font-semibold text-xl">
                      <span className="">Alby Hub</span>
                    </Link>

                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger>
                          <ExternalLink
                            to="https://getalby.com/hub_deployment/edit" // TODO: link to general update page
                            className="font-semibold text-xl"
                          >
                            <span className="text-xs flex items-center text-muted-foreground">
                              {info?.version}&nbsp;
                              {upToDate ? (
                                <ShieldCheckIcon className="w-4 h-4" />
                              ) : (
                                <ShieldAlertIcon className="w-4 h-4" />
                              )}
                            </span>
                          </ExternalLink>
                        </TooltipTrigger>
                        <TooltipContent>
                          {upToDate ? (
                            <p>Alby Hub is up to date!</p>
                          ) : (
                            <p>Alby Hub {info?.latestVersion} available!</p>
                          )}
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                  <MainMenuContent />
                </nav>
              </div>
              <div className="flex flex-col">
                <SidebarHint />
                <MainNavSecondary />
                <div className="flex h-14 items-center px-4 lg:h-[60px] lg:px-6 gap-3 border-t border-border justify-between">
                  <div className="grid grid-flow-col gap-2">
                    <UserAvatar className="h-8 w-8" />
                    <Link
                      to="#"
                      className="font-semibold text-lg whitespace-nowrap overflow-hidden text-ellipsis"
                    >
                      {albyMe?.name || albyMe?.email}
                    </Link>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon">
                        <EllipsisVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <UserMenuContent />
                  </DropdownMenu>
                </div>
              </div>
            </div>
          </div>
          <main className="flex flex-col">
            <header className="md:hidden flex h-14 items-center gap-4 border-b bg-muted/40 px-4 lg:h-[60px] lg:px-6 justify-between">
              <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
                <SheetTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon"
                    className="shrink-0 md:hidden"
                  >
                    <Menu className="h-5 w-5" />
                    <span className="sr-only">Toggle navigation menu</span>
                  </Button>
                </SheetTrigger>
                <SheetContent
                  side="left"
                  className="flex flex-col justify-between max-h-screen"
                >
                  <nav className="grid gap-2 text-lg font-medium">
                    <div className="p-3 ">
                      <Link to="/" className="font-semibold text-xl">
                        <span className="">Alby Hub</span>
                      </Link>
                    </div>
                    <MainMenuContent />
                  </nav>
                  <div className="align-bottom">
                    <SidebarHint />
                    <MainNavSecondary />
                  </div>
                </SheetContent>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Link
                      to="#"
                      className="grid grid-flow-col gap-2 font-semibold text-lg whitespace-nowrap overflow-hidden text-ellipsis"
                    >
                      <UserAvatar className="h-8 w-8" />
                    </Link>
                  </DropdownMenuTrigger>
                  <UserMenuContent />
                </DropdownMenu>
              </Sheet>
            </header>
            <div className="flex flex-1 flex-col gap-4 p-4 lg:gap-6 lg:p-8">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </>
  );
}

const MenuItem = ({
  to,
  children,
  disabled = false,
  onClick,
}: {
  to: string;
  children: React.ReactNode | string;
  disabled?: boolean;
  onClick?: (e: React.MouseEvent<HTMLElement>) => void;
}) => (
  <NavLink
    to={to}
    onClick={(e) => {
      if (disabled) {
        e.preventDefault();
      }
      if (onClick) {
        onClick(e);
      }
    }}
    className={({ isActive }) =>
      cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary",
        disabled && "cursor-not-allowed",
        !disabled && isActive ? "bg-muted" : ""
      )
    }
  >
    {children}
  </NavLink>
);
