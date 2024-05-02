import {
  Cable,
  Menu,
  SendToBack,
  Settings,
  Store,
  Wallet
} from "lucide-react";
import { ModeToggle } from "src/components/ui/mode-toggle";

import { CaretUpIcon } from "@radix-ui/react-icons";
import React from "react";
import {
  Link,
  NavLink,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { Button } from "src/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "src/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetTrigger } from "src/components/ui/sheet";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";
import { request } from "src/utils/request";

export default function AppLayout() {
  const { data: albyMe } = useAlbyMe();
  const { data: csrf } = useCSRF();
  const { toast } = useToast();
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);
  const location = useLocation();
  const navigate = useNavigate();

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

    navigate("/", { replace: true });
    toast({ title: "You are now logged out." });
  }, [csrf, navigate, toast]);

  function UserMenuContent() {
    return (
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>My Account</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem>
            <a
              href="https://getalby.com/settings/alby_page"
              target="_blank"
              rel="noreferer noopener"
              className="w-full"
            >
              Profile
            </a>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={logout}>Log out</DropdownMenuItem>
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
    const { data: info } = useInfo();
    return (
      <nav className="grid items-start p-2 text-sm font-medium lg:px-4">
        <div className="px-3 py-2 mb-5">
          <ModeToggle />
        </div>
        {(info?.backendType === "LDK" ||
          info?.backendType === "GREENLIGHT") && (
            <MenuItem to="/channels">
              <SendToBack className="h-4 w-4" />
              Channels
            </MenuItem>
          )}
        <MenuItem to="/settings">
          <Settings className="h-4 w-4" />
          Settings
        </MenuItem>
      </nav>
    );
  }

  return (
    <div className="font-sans grid min-h-screen w-full md:grid-cols-[220px_1fr] lg:grid-cols-[280px_1fr]">
      <div className="hidden border-r bg-muted/40 md:block">
        <div className="flex h-full max-h-screen flex-col gap-2">
          <div className="flex-1">
            <nav className="grid items-start px-2 py-2 text-sm font-medium lg:px-4">
              <div className="p-3 ">
                <Link to="/" className="font-semibold text-xl">
                  <span className="">Alby Hub</span>
                </Link>
              </div>
              <MainMenuContent />
            </nav>
          </div>
          <div className="flex flex-col">
            <MainNavSecondary />
            <div className="flex h-14 items-center px-4 lg:h-[60px] lg:px-6 gap-3 border-t border-border justify-between">
              <div className="grid grid-flow-col gap-2">
                <Avatar className="h-8 w-8">
                  <AvatarImage src={albyMe?.avatar} alt="@satoshi" />
                  <AvatarFallback>SN</AvatarFallback>
                </Avatar>
                <Link
                  to="#"
                  className="font-semibold text-lg whitespace-nowrap overflow-hidden text-ellipsis"
                >
                  {albyMe?.name ?? albyMe?.email}
                </Link>
              </div>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <CaretUpIcon />
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
                <MainNavSecondary />
              </div>
            </SheetContent>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Link
                  to="#"
                  className="grid grid-flow-col gap-2 font-semibold text-lg whitespace-nowrap overflow-hidden text-ellipsis"
                >
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={albyMe?.avatar} alt="@satoshi" />
                    <AvatarFallback>SN</AvatarFallback>
                  </Avatar>
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
  );
}

const MenuItem = ({
  to,
  children,
  disabled = false,
}: {
  to: string;
  children: React.ReactNode | string;
  disabled?: boolean;
}) => (
  <NavLink
    to={to}
    onClick={(e) => {
      if (disabled) e.preventDefault();
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

MenuItem;
