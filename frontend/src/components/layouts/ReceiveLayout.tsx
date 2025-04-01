import { ArrowDownIcon } from "lucide-react";
import { Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import BalanceCard from "src/components/BalanceCard";
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
      <AppHeader title="Receive" />
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          <Outlet />
        </div>
        {hasChannelManagement && (
          <BalanceCard
            balance={balances.lightning.totalReceivable}
            title="Receiving Capacity"
            buttonTitle="Increase"
            buttonLink="/channels/incoming"
            BalanceCardIcon={ArrowDownIcon}
            hasChannelManagement
          />
        )}
      </div>
    </div>
  );
}
