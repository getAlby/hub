import React from "react";
import { NavLink, Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { buttonVariants } from "src/components/ui/button";

import { useInfo } from "src/hooks/useInfo";

import { cn } from "src/lib/utils";

export default function SettingsLayout() {
  const { data: info, hasMnemonic, hasNodeBackup } = useInfo();

  return (
    <>
      <AppHeader
        title="Settings"
        description=""
        breadcrumb={false}
        contentRight={
          info?.version && (
            <p className="text-sm text-muted-foreground">{info.version}</p>
          )
        }
      />
      <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-12 lg:space-y-0">
        <aside className="lg:-mx-4 lg:w-1/5">
          <nav className="flex flex-wrap lg:flex-col -space-x-1 lg:space-x-0 lg:space-y-1">
            <MenuItem to="/settings">Theme</MenuItem>
            <MenuItem to="/settings/change-unlock-password">
              Unlock Password
            </MenuItem>
            {(hasMnemonic || hasNodeBackup) && (
              <MenuItem to="/settings/backup">Backup</MenuItem>
            )}
            {info?.albyAccountConnected && (
              <MenuItem to="/settings/alby-account">Your Alby Account</MenuItem>
            )}
            {info && !info.albyAccountConnected && (
              <MenuItem to="/alby/account">Alby Account</MenuItem>
            )}
            <MenuItem to="/settings/developer">Developer</MenuItem>
            <MenuItem to="/settings/debug-tools">Debug Tools</MenuItem>
            <MenuItem to="/settings/shutdown">Shutdown</MenuItem>
          </nav>
        </aside>
        <div className="flex-1 lg:max-w-2xl">
          <div className="grid gap-8">
            <Outlet />
          </div>
        </div>
      </div>
    </>
  );
}

const MenuItem = ({
  to,
  children,
}: {
  to: string;
  children: React.ReactNode | string;
}) => (
  <>
    <NavLink
      end
      to={to}
      className={({ isActive }) =>
        cn(
          buttonVariants({ variant: "ghost" }),
          isActive
            ? "bg-muted hover:bg-muted"
            : "hover:bg-transparent hover:underline",
          "justify-start"
        )
      }
    >
      {children}
    </NavLink>
  </>
);

MenuItem;
