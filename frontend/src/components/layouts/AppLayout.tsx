import { Outlet } from "react-router-dom";

import { AppSidebar } from "src/components/AppSidebar";
import { Banner } from "src/components/Banner";
import { SidebarInset, SidebarProvider } from "src/components/ui/sidebar";
import { useBanner } from "src/hooks/useBanner";
import { useInfo } from "src/hooks/useInfo";
import { useNotifyReceivedPayments } from "src/hooks/useNotifyReceivedPayments";
import { useRemoveSuccessfulChannelOrder } from "src/hooks/useRemoveSuccessfulChannelOrder";
import { cn } from "src/lib/utils";

export default function AppLayout() {
  const { data: info } = useInfo();
  const { showBanner, dismissBanner } = useBanner();

  useRemoveSuccessfulChannelOrder();
  useNotifyReceivedPayments();

  if (!info) {
    return null;
  }

  return (
    <>
      <div
        className={cn(
          "font-sans min-h-screen w-full flex flex-col",
          showBanner
            ? "[--header-height:calc(theme(spacing.9))]" // Banner height is 36px when visible (sidebar hidden on <md width)
            : "[--header-height:0]"
        )}
      >
        <SidebarProvider className="flex flex-col">
          {showBanner && <Banner onDismiss={dismissBanner} />}
          <div className="flex flex-1">
            <AppSidebar />
            <SidebarInset>
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
    </>
  );
}
