import React, { useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router";
import AppHeader from "src/components/AppHeader";

import { useInfo } from "src/hooks/useInfo";

import {
  ArrowRightLeftIcon,
  BugIcon,
  CloudBackupIcon,
  CodeIcon,
  FingerprintIcon,
  InfoIcon,
  KeyRoundIcon,
  type LucideIcon,
  PowerIcon,
  SlidersHorizontalIcon,
  UserIcon,
} from "lucide-react";
import { toast } from "sonner";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Separator } from "src/components/ui/separator";
import { cn } from "src/lib/utils";
import { request } from "src/utils/request";

export default function SettingsLayout() {
  const {
    data: info,
    mutate: refetchInfo,
    hasMnemonic,
    hasNodeBackup,
  } = useInfo();
  const navigate = useNavigate();
  const [shuttingDown, setShuttingDown] = useState(false);

  const shutdown = React.useCallback(async () => {
    setShuttingDown(true);
    try {
      await request("/api/stop", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      await refetchInfo();
      setShuttingDown(false);
      navigate("/", { replace: true });
      toast("Your node has been turned off.");
    } catch (error) {
      console.error(error);
      toast.error("Failed to shutdown node", {
        description: "" + error,
      });
    }
  }, [navigate, refetchInfo]);

  return (
    <>
      <AppHeader
        title="Settings"
        breadcrumb={false}
        contentRight={
          <div className="flex items-center gap-4">
            <div className="font-medium slashed-zero text-muted-foreground text-sm">
              {info?.version}
            </div>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <LoadingButton
                  variant="destructive"
                  size="icon"
                  loading={shuttingDown}
                >
                  {!shuttingDown && <PowerIcon className="size-4" />}
                </LoadingButton>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    Do you want to turn off your Alby Hub?
                  </AlertDialogTitle>
                  <AlertDialogDescription>
                    This will turn off your Alby Hub and make your node offline.
                    You won't be able to send or receive bitcoin until you
                    unlock it.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction onClick={shutdown}>
                    Continue
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        }
      />

      <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-4 lg:space-y-0 h-full">
        <aside className="flex flex-col justify-between lg:w-1/5">
          <nav className="flex overflow-x-auto pb-2 gap-1 lg:flex-col lg:overflow-x-visible lg:pb-0 lg:gap-0 lg:space-y-0.5">
            <MenuItem to="/settings" icon={SlidersHorizontalIcon}>
              General
            </MenuItem>

            <NavGroup label="Security">
              <MenuItem
                to="/settings/change-unlock-password"
                icon={KeyRoundIcon}
              >
                Unlock Password
              </MenuItem>
              {info?.autoUnlockPasswordSupported && (
                <MenuItem to="/settings/auto-unlock" icon={FingerprintIcon}>
                  Auto Unlock
                </MenuItem>
              )}
            </NavGroup>

            {(hasMnemonic || hasNodeBackup) && (
              <NavGroup label="Data">
                {hasMnemonic && (
                  <MenuItem to="/settings/backup" icon={CloudBackupIcon}>
                    Backup
                  </MenuItem>
                )}
                {hasNodeBackup && (
                  <MenuItem
                    to="/settings/node-migrate"
                    icon={ArrowRightLeftIcon}
                  >
                    Migrate Alby Hub
                  </MenuItem>
                )}
              </NavGroup>
            )}

            <NavGroup label="Account">
              {info?.albyAccountConnected && (
                <MenuItem to="/settings/alby-account" icon={UserIcon}>
                  Your Alby Account
                </MenuItem>
              )}
              {info && !info.albyAccountConnected && (
                <MenuItem to="/alby/account" icon={UserIcon}>
                  Alby Account
                </MenuItem>
              )}
            </NavGroup>

            <NavGroup label="Advanced">
              <MenuItem to="/settings/developer" icon={CodeIcon}>
                Developer
              </MenuItem>
              <MenuItem to="/settings/debug-tools" icon={BugIcon}>
                Debug Tools
              </MenuItem>
              <MenuItem to="/settings/about" icon={InfoIcon}>
                About
              </MenuItem>
            </NavGroup>
          </nav>
        </aside>
        <Separator orientation="vertical" className="hidden lg:block" />
        <div className="flex-1 lg:max-w-2xl">
          <div className="grid gap-6">
            <Outlet />
          </div>
        </div>
      </div>
    </>
  );
}

const NavGroup = ({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) => (
  <div className="contents lg:block lg:pt-4 lg:space-y-0.5">
    <span className="hidden lg:block px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground/60">
      {label}
    </span>
    {children}
  </div>
);

export const MenuItem = ({
  to,
  icon: Icon,
  children,
}: {
  to: string;
  icon?: LucideIcon;
  children: React.ReactNode | string;
}) => (
  <NavLink
    end
    to={to}
    className={({ isActive }) =>
      cn(
        "relative flex items-center gap-2 whitespace-nowrap rounded-md px-3 py-2 text-sm transition-colors text-foreground",
        isActive ? "bg-muted font-medium" : "hover:bg-muted/50"
      )
    }
  >
    {() => (
      <>
        {Icon && <Icon className="size-4 shrink-0 text-foreground" />}
        {children}
      </>
    )}
  </NavLink>
);
