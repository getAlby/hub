import { compare } from "compare-versions";
import {
  BoxIcon,
  CircleHelpIcon,
  EllipsisVerticalIcon,
  HomeIcon,
  LayoutGridIcon,
  LockIcon,
  MenuIcon,
  Plug2Icon,
  PlugZapIcon,
  SettingsIcon,
  ShieldAlertIcon,
  ShieldCheckIcon,
  SparklesIcon,
  User2Icon,
  WalletIcon,
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
import { UpgradeDialog } from "src/components/UpgradeDialog";
import UserAvatar from "src/components/UserAvatar";
import { AlbyHubLogo } from "src/components/icons/AlbyHubLogo";
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
import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useHealthCheck } from "src/hooks/useHealthCheck";
import { useInfo } from "src/hooks/useInfo";
import { useNotifyReceivedPayments } from "src/hooks/useNotifyReceivedPayments";
import { useRemoveSuccessfulChannelOrder } from "src/hooks/useRemoveSuccessfulChannelOrder";
import { deleteAuthToken } from "src/lib/auth";
import { cn } from "src/lib/utils";
import { isHttpMode } from "src/utils/isHttpMode";
import { openLink } from "src/utils/openLink";
import ExternalLink from "../ExternalLink";

export default function AppLayout() {
  const { data: albyMe } = useAlbyMe();

  const { data: info, mutate: refetchInfo } = useInfo();
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  useRemoveSuccessfulChannelOrder();
  useNotifyReceivedPayments();

  const _isHttpMode = isHttpMode();

  React.useEffect(() => {
    setMobileMenuOpen(false);
  }, [location]);

  const logout = React.useCallback(async () => {
    deleteAuthToken();
    await refetchInfo();

    if (_isHttpMode) {
      window.location.href = "/logout";
    } else {
      navigate("/", { replace: true });
    }
  }, [_isHttpMode, navigate, refetchInfo]);

  if (!info) {
    return null;
  }

  function UserMenuContent() {
    return (
      <DropdownMenuContent align="end">
        <DropdownMenuGroup>
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
          {info?.albyAccountConnected && (
            <DropdownMenuItem>
              <Link
                to="/settings/alby-account"
                className="w-full flex flex-row items-center gap-2"
              >
                <User2Icon className="w-4 h-4" />
                <p>Alby Account Settings</p>
              </Link>
            </DropdownMenuItem>
          )}
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        {_isHttpMode && (
          <DropdownMenuItem
            onClick={logout}
            className="w-full flex flex-row items-center gap-2 cursor-pointer"
          >
            <LockIcon className="w-4 h-4" />
            <p>Lock Alby Hub</p>
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    );
  }

  function MainMenuContent() {
    return (
      <>
        <MenuItem to="/home">
          <HomeIcon className="h-4 w-4" />
          Home
        </MenuItem>
        <MenuItem to="/wallet">
          <WalletIcon className="h-4 w-4" />
          Wallet
        </MenuItem>
        <MenuItem to="/appstore">
          <LayoutGridIcon className="h-4 w-4" />
          App Store
        </MenuItem>
        <MenuItem to="/apps">
          <Plug2Icon className="h-4 w-4" />
          Connections
        </MenuItem>
      </>
    );
  }

  function MainNavSecondary() {
    const { hasChannelManagement } = useInfo();
    return (
      <nav className="grid items-start md:px-4 md:py-2 text-sm font-medium">
        {hasChannelManagement && (
          <MenuItem to="/channels">
            <BoxIcon className="h-4 w-4" />
            Node
          </MenuItem>
        )}
        <MenuItem to="/settings">
          <SettingsIcon className="h-4 w-4" />
          Settings
        </MenuItem>

        <MenuItem
          to="/"
          onClick={(e) => {
            openLink("https://support.getalby.com");
            openLink("https://support.getalby.com");
            e.preventDefault();
          }}
        >
          <CircleHelpIcon className="h-4 w-4" />
          Help
        </MenuItem>
        <UpgradeDialog>
          <div className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-accent-foreground">
            <SparklesIcon className="h-4 w-4" />
            Upgrade
          </div>
        </UpgradeDialog>
      </nav>
    );
  }

  return (
    <>
      <div className="font-sans min-h-screen w-full flex flex-col">
        <div className="flex-1 h-full md:grid md:grid-cols-[280px_minmax(0,1fr)]">
          <div className="hidden border-r bg-muted/40 md:block">
            <div className="flex h-full max-h-screen flex-col gap-2 fixed w-[280px] z-10 top-0 overflow-y-auto">
              <div className="flex-1">
                <nav className="grid items-start px-4 py-2 text-sm font-medium">
                  <div className="p-3 flex justify-between items-center mt-2 mb-6">
                    <Link to="/">
                      <AlbyHubLogo className="text-foreground" />
                    </Link>
                    <AppVersion />
                    <HealthIndicator />
                  </div>
                  <MainMenuContent />
                </nav>
              </div>
              <div className="flex flex-col">
                <SidebarHint />
                <MainNavSecondary />
                <div className="flex h-14 items-center px-4 gap-3 border-t border-border justify-between">
                  {info.albyAccountConnected ? (
                    <div className="grid grid-flow-col gap-2 items-center">
                      <UserAvatar className="h-8 w-8" />
                      <Link
                        to="#"
                        className="font-semibold text-lg whitespace-nowrap overflow-hidden text-ellipsis"
                      >
                        {albyMe?.name || albyMe?.email}
                      </Link>
                    </div>
                  ) : (
                    <div></div>
                  )}
                  <DropdownMenu modal={false}>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon">
                        <EllipsisVerticalIcon className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <UserMenuContent />
                  </DropdownMenu>
                </div>
              </div>
            </div>
          </div>
          <main className="flex flex-col">
            <header className="md:hidden fixed w-full top-0 z-50 flex h-14 items-center gap-4 border-b bg-muted/40 backdrop-blur px-4 lg:h-[60px] lg:px-6 justify-between">
              <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
                <SheetTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon"
                    className="shrink-0 md:hidden"
                  >
                    <MenuIcon className="h-5 w-5" />
                    <span className="sr-only">Toggle navigation menu</span>
                  </Button>
                </SheetTrigger>
                <SheetContent
                  side="left"
                  className="flex flex-col justify-between max-h-screen px-4"
                >
                  <nav className="grid text-sm font-medium">
                    <div className="p-3 pr-0 flex justify-between items-center">
                      <Link to="/">
                        <AlbyHubLogo className="text-foreground" />
                      </Link>
                      <div className="mr-1 flex gap-2 items-center justify-center">
                        <AppVersion />
                        <HealthIndicator />
                      </div>
                    </div>
                    <MainMenuContent />
                  </nav>
                  <div className="align-bottom">
                    <SidebarHint />
                    <MainNavSecondary />
                  </div>
                </SheetContent>
                <DropdownMenu modal={false}>
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
            <div className="flex flex-1 flex-col gap-4 p-4 mt-14 md:mt-0 lg:gap-6 lg:p-8">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </>
  );
}

function AppVersion() {
  const { data: albyInfo } = useAlbyInfo();
  const { data: info } = useInfo();
  if (!info || !albyInfo) {
    return null;
  }

  const upToDate =
    info.version &&
    info.version.startsWith("v") &&
    compare(info.version.substring(1), albyInfo.hub.latestVersion, ">=");

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <ExternalLink
            to={`https://getalby.com/update/hub?version=${info.version}`}
            className="font-semibold text-xl"
          >
            <span className="text-xs flex items-center text-muted-foreground">
              {info.version && <>{info.version}&nbsp;</>}
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
            <div>
              <p className="font-semibold">
                Alby Hub {albyInfo.hub.latestVersion} available!
              </p>
              <p className="mt-2 max-w-xs whitespace-pre-wrap">
                {albyInfo.hub.latestReleaseNotes}
              </p>
            </div>
          )}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

function HealthIndicator() {
  const { data: health } = useHealthCheck();
  if (!health) {
    return null;
  }

  const ok = !health.alarms?.length;

  return (
    <Link to="/channels?healthcheck=true">
      <div
        className={cn(
          "w-2 h-2 rounded-full",
          ok ? "bg-green-300" : "bg-destructive"
        )}
      />
    </Link>
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
        "flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-accent-foreground",
        disabled && "cursor-not-allowed",
        !disabled && isActive ? "bg-muted" : ""
      )
    }
  >
    {children}
  </NavLink>
);
