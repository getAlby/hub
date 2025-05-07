import { Outlet } from "react-router-dom";

import { AppSidebar } from "src/components/AppSidebar";
import { SidebarInset, SidebarProvider } from "src/components/ui/sidebar";
import { UpdateBanner } from "src/components/UpdateBanner";
import { useInfo } from "src/hooks/useInfo";
import { useNotifyReceivedPayments } from "src/hooks/useNotifyReceivedPayments";
import { useRemoveSuccessfulChannelOrder } from "src/hooks/useRemoveSuccessfulChannelOrder";

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
        <UpdateBanner />
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset>
            <div className="flex flex-1 flex-col gap-4 p-4">
              <Outlet />
            </div>
          </SidebarInset>
        </SidebarProvider>
      </div>
    </>
  );
}
