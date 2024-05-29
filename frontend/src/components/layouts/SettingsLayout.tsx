import { ExternalLink, Power } from "lucide-react";
import React from "react";
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
import { Button, buttonVariants } from "src/components/ui/button";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";

import { cn } from "src/lib/utils";
import { request } from "src/utils/request";

export default function SettingsLayout() {
  const { data: csrf } = useCSRF();
  const { mutate: refetchInfo } = useInfo();
  const navigate = useNavigate();
  const { toast } = useToast();

  const shutdown = React.useCallback(async () => {
    if (!csrf) {
      throw new Error("csrf not loaded");
    }

    await request("/api/stop", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
    });

    await refetchInfo();
    navigate("/", { replace: true });
    toast({ title: "Your node has been turned off." });
  }, [csrf, navigate, refetchInfo, toast]);

  const { data: info } = useInfo();

  return (
    <>
      <AppHeader
        title="Settings"
        description="Manage your Alby Hub settings."
        contentRight={
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive" size="icon">
                <Power className="w-4 h-4" />
              </Button>
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
            { (info?.backendType === "LDK" || info?.backendType === "BREEZ" || info?.backendType === "GREENLIGHT") && (
              <MenuItem to="/settings/key-backup">Key Backup</MenuItem>
            )}
            {info?.backendType === "LDK" && (
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
