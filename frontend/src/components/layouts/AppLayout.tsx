import {
  Cable,
  CircleHelp,
  LayoutGrid,
  MessageCircle,
  SendToBack,
  Settings,
  ShieldCheck,
  Wallet,
} from "lucide-react";
import { ModeToggle } from "src/components/ui/mode-toggle";

import { Link, NavLink, Outlet } from "react-router-dom";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { cn } from "src/lib/utils";

export default function AppLayout() {
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
              <MenuItem to="/wallet">
                <Wallet className="h-4 w-4" />
                Wallet
              </MenuItem>
              <MenuItem to="/apps">
                <Cable className="h-4 w-4" />
                Apps
              </MenuItem>
              <MenuItem to="/appstore">
                <LayoutGrid className="h-4 w-4" />
                Store
              </MenuItem>
              <MenuItem to="/permissions" disabled>
                <ShieldCheck className="h-4 w-4" />
                Permissions
              </MenuItem>
            </nav>
          </div>
          <div className="flex flex-col">
            <nav className="grid items-start p-2 text-sm font-medium lg:px-4">
              <div className="px-3 py-2 mb-5">
                <ModeToggle />
              </div>
              <MenuItem to="/channels">
                <SendToBack className="h-4 w-4" />
                Channels
              </MenuItem>

              <MenuItem to="/settings">
                <Settings className="h-4 w-4" />
                Settings
              </MenuItem>
              <MenuItem to="/help" disabled>
                <CircleHelp className="h-4 w-4" />
                Help
              </MenuItem>
              <MenuItem to="feedback" disabled>
                <MessageCircle className="h-4 w-4" />
                Leave Feedback
              </MenuItem>
            </nav>
            <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6 gap-3 border-t border-border">
              <Avatar className="h-8 w-8">
                <AvatarImage
                  src="https://github.com/shadcn.png"
                  alt="@satoshi"
                />
                <AvatarFallback>SN</AvatarFallback>
              </Avatar>
              <Link
                to="#"
                className="flex items-center gap-2 font-semibold text-lg cursor-not-allowed"
              >
                Satoshi
              </Link>
            </div>
          </div>
        </div>
      </div>
      <main className="flex flex-1 flex-col gap-4 p-4 lg:gap-6 lg:p-8">
        <Outlet />
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
  <>
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
  </>
);

MenuItem;
