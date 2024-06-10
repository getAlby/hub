import { ExternalLink, Power } from "lucide-react";
import React, { useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
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
import { buttonVariants } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";

import { cn } from "src/lib/utils";
import { request } from "src/utils/request";

export default function SettingsLayout() {
  const { data: csrf } = useCSRF();
  const { mutate: refetchInfo, hasMnemonic, hasNodeBackup } = useInfo();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [shuttingDown, setShuttingDown] = useState(false);

  const shutdown = React.useCallback(async () => {
    if (!csrf) {
      throw new Error("csrf not loaded");
    }

    setShuttingDown(true);
    try {
      await request("/api/stop", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
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
  }, [csrf, navigate, refetchInfo, toast]);

  return (
    <>
      <AppHeader
        title="Settings"
        description="Manage your Alby Hub settings."
        contentRight={
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <LoadingButton
                variant="destructive"
                size="icon"
                loading={shuttingDown}
              >
                {!shuttingDown && <Power className="w-4 h-4" />}
              </LoadingButton>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>
                  Do you want to turn off Alby Hub?
                </AlertDialogTitle>
                <AlertDialogDescription>
                  This will turn off your Alby Hub and make your node offline.
                  You won't be able to send or receive bitcoin until you unlock
                  it.
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
        }
      />
      <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-12 lg:space-y-0">
        <aside className="lg:-mx-4 lg:w-1/5">
          <nav className="flex space-x-2 lg:flex-col lg:space-x-0 lg:space-y-1">
            <MenuItem to="/settings">General</MenuItem>
            <MenuItem to="/settings/change-unlock-password">
              Unlock Password
            </MenuItem>
            {hasMnemonic && (
              <MenuItem to="/settings/key-backup">Key Backup</MenuItem>
            )}
            {hasNodeBackup && (
              <MenuItem to="/settings/node-backup">Node Backup</MenuItem>
            )}
            <MenuItem to="/debug-tools">
              Debug Tools
              <ExternalLink className="w-4 h-4 ml-2" />
            </MenuItem>
          </nav>
        </aside>
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
