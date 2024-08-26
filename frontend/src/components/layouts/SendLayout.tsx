import { ArrowUp } from "lucide-react";
import { Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import BalanceCard from "src/components/BalanceCard";
import Loading from "src/components/Loading";
import { useBalances } from "src/hooks/useBalances";
import { useChannels } from "src/hooks/useChannels";

import { useInfo } from "src/hooks/useInfo";

export default function Send() {
  const { hasChannelManagement } = useInfo();
  const { data: balances } = useBalances();
  const { data: channels } = useChannels();

  if (!balances || !channels) {
    return <Loading />;
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Send"
        description="Pay a lightning invoice created by any bitcoin lightning wallet"
      />
      <div className="flex gap-12 w-full">
        <div className="w-full max-w-lg">
          <Outlet />
        </div>
        <BalanceCard
          balance={
            balances
              ? new Intl.NumberFormat(undefined, {}).format(
                  Math.floor(balances.lightning.totalSpendable / 1000)
                )
              : ""
          }
          title="Spending Balance"
          buttonTitle="Top Up"
          buttonLink="/channels/outgoing"
          BalanceCardIcon={ArrowUp}
          hasChannelManagement={!!hasChannelManagement}
        />
      </div>
    </div>
  );
}
