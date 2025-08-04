import React, { useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { buttonVariants } from "src/components/ui/button";

import { useInfo } from "src/hooks/useInfo";

import { PowerIcon } from "lucide-react";
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
import { LoadingButton } from "src/components/ui/loading-button";
import { Separator } from "src/components/ui/separator";
import { useToast } from "src/components/ui/use-toast";
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
  const { toast } = useToast();
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
      toast({ title: "Your node has been turned off." });
    } catch (error) {
      console.error(error);
      toast({
        title: "Failed to shutdown node: " + error,
        variant: "destructive",
      });
    }
  }, [navigate, refetchInfo, toast]);

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
          <nav className="flex flex-wrap lg:flex-col lg:space-y-1">
            <MenuItem to="/settings">General</MenuItem>
            {info?.autoUnlockPasswordSupported && (
              <MenuItem to="/settings/auto-unlock">Auto Unlock</MenuItem>
            )}
            <MenuItem to="/settings/change-unlock-password">
              Unlock Password
            </MenuItem>
            {hasMnemonic && <MenuItem to="/settings/backup">Backup</MenuItem>}
            {hasNodeBackup && (
              <MenuItem to="/settings/node-migrate">Migrate Alby Hub</MenuItem>
            )}
            {info?.albyAccountConnected && (
              <MenuItem to="/settings/alby-account">Your Alby Account</MenuItem>
            )}
            {info && !info.albyAccountConnected && (
              <MenuItem to="/alby/account">Alby Account</MenuItem>
            )}
            <MenuItem to="/settings/developer">Developer</MenuItem>
            <MenuItem to="/settings/debug-tools">Debug Tools</MenuItem>
            <MenuItem to="/settings/about">About</MenuItem>
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
