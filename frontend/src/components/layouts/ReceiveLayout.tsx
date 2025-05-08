import { Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";

export default function ReceiveLayout() {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();

  if (!balances || !channels) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Receive"
        contentRight={
          hasChannelManagement && (
            <div className="md:flex md:items-center md:gap-4">
              <span className="text-muted-foreground">Receive Limit:</span>
              <div className="balance sensitive slashed-zero">
                {new Intl.NumberFormat().format(
                  Math.floor(balances.lightning.totalReceivable / 1000)
                )}{" "}
                sats
              </div>
            </div>
          )
        }
      />
      <Outlet />
    </div>
  );
}
