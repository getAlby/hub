import React from "react";
import { Outlet, useMatches } from "react-router-dom";

import { AppSidebar } from "src/components/AppSidebar";
import { Banner } from "src/components/Banner";
import { CommandPalette } from "src/components/CommandPalette";
import { SidebarInset, SidebarProvider } from "src/components/ui/sidebar";
import {
  CommandPaletteProvider,
  useCommandPaletteContext,
} from "src/contexts/CommandPaletteContext";
import { useBanner } from "src/hooks/useBanner";
import { useInfo } from "src/hooks/useInfo";
import { useNotifyReceivedPayments } from "src/hooks/useNotifyReceivedPayments";
import { useRemoveSuccessfulChannelOrder } from "src/hooks/useRemoveSuccessfulChannelOrder";
import { cn } from "src/lib/utils";

function AppLayoutInner() {
  const { data: info } = useInfo();
  const { showBanner, dismissBanner } = useBanner();
  const { open, setOpen } = useCommandPaletteContext();

  useRemoveSuccessfulChannelOrder();
  useNotifyReceivedPayments();

  // Update document.title and the history entry when the route changes.
  // This ensures the browser's history entries include a proper title
  // (fixes: back gesture / long-press back showing empty/incorrect titles).
  const matches = useMatches();
  React.useEffect(() => {
    try {
      // Attempt to derive a title from route handles (crumb) or matched route
      // fallback to app name. Use a small typed helper instead of `any` to
      // satisfy lint rules.
      const getCrumbFromHandle = (
        handle: unknown
      ): string | string[] | null => {
        if (handle && typeof handle === "object") {
          const h = handle as { crumb?: unknown };
          if (typeof h.crumb === "function") {
            try {
              return (h.crumb as () => string | string[])();
            } catch (err) {
              return null;
            }
          }
        }
        return null;
      };

      const crumbTitle =
        matches
          .map((m) => getCrumbFromHandle(m.handle))
          .filter(Boolean)
          .pop() || "Alby Hub";

      const title = Array.isArray(crumbTitle)
        ? crumbTitle.join(" - ")
        : crumbTitle;

      // Set document title
      document.title = title as string;

      // Replace current history state to include title in state (some browsers show
      // history entry title from state). We keep the existing state but add _title.
      try {
        const state =
          history.state && typeof history.state === "object"
            ? { ...history.state }
            : {};
        if (state && state._title !== title) {
          state._title = title;
          history.replaceState(state, title as string, window.location.href);
        }
      } catch (err) {
        // ignore replaceState errors in weird environments
        // eslint-disable-next-line no-console
        console.debug("history.replaceState failed", err);
      }
    } catch (err) {
      console.error("Failed to compute page title", err);
    }
  }, [matches]);

  if (!info) {
    return null;
  }

  return (
    <>
      <div
        className={cn(
          "font-sans min-h-screen w-full flex flex-col",
          showBanner
            ? "[--header-height:calc(--spacing(9))]" // Banner height is 36px when visible (sidebar hidden on <md width)
            : "[--header-height:0]"
        )}
      >
        <SidebarProvider className="flex flex-col">
          {showBanner && <Banner onDismiss={dismissBanner} />}
          <div className="flex flex-1">
            <AppSidebar />
            <SidebarInset className="min-w-0">
              <div
                className={cn(
                  "flex flex-1 flex-col gap-4 p-4",
                  showBanner && "mt-14 md:mt-9" // Banner height is 36px with 1 line (>=md width) and 56px with 2 lines (<md width)
                )}
              >
                <Outlet />
              </div>
            </SidebarInset>
          </div>
        </SidebarProvider>
      </div>
      <CommandPalette open={open} onOpenChange={setOpen} />
    </>
  );
}

export default function AppLayout() {
  return (
    <CommandPaletteProvider>
      <AppLayoutInner />
    </CommandPaletteProvider>
  );
}
