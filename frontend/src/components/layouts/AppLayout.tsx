import { compare } from "compare-versions";
import { ShieldAlertIcon, ShieldCheckIcon } from "lucide-react";

import React from "react";
import { NavLink, Outlet } from "react-router-dom";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";

import { AppSidebar } from "src/components/AppSidebar";
import { SidebarInset, SidebarProvider } from "src/components/ui/sidebar";
import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useInfo } from "src/hooks/useInfo";
import { useNotifyReceivedPayments } from "src/hooks/useNotifyReceivedPayments";
import { useRemoveSuccessfulChannelOrder } from "src/hooks/useRemoveSuccessfulChannelOrder";
import { cn } from "src/lib/utils";
import ExternalLink from "../ExternalLink";

export default function AppLayout() {
  const { data: info } = useInfo();
  useRemoveSuccessfulChannelOrder();
  useNotifyReceivedPayments();

  if (!info) {
    return null;
  }

  return (
    <>
      <div className="font-sans min-h-screen w-full flex flex-col">
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset>
            <div className="flex flex-1 flex-col gap-4 p-4">
              <Outlet />
            </div>
          </SidebarInset>
        </SidebarProvider>
        {/* <AppVersion /> */}
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
