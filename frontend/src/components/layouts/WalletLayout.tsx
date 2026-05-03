import {
  ArrowDownUpIcon,
  CalendarSyncIcon,
  CreditCardIcon,
} from "lucide-react";
import { Outlet } from "react-router";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LinkButton } from "src/components/ui/custom/link-button";
import { WalletActionsMenu } from "src/components/WalletActionsMenu";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import { useSyncWallet } from "src/hooks/useSyncWallet";

export default function WalletLayout() {
  useSyncWallet();
  const { data: info, hasChannelManagement } = useInfo();
  const { data: balances } = useBalances(true);

  if (!info || !balances) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader
        title="Wallet"
        pageTitle="Wallet"
        description=""
        contentRight={
          <div className="flex items-center gap-1 sm:gap-2">
            {hasChannelManagement && (
              <LinkButton
                to="/wallet/swap"
                variant="ghost"
                size="sm"
                className="hidden sm:inline-flex"
              >
                <ArrowDownUpIcon />
                Swap
              </LinkButton>
            )}
            <LinkButton
              to="/internal-apps/zapplanner"
              variant="ghost"
              size="sm"
              className="hidden sm:inline-flex"
            >
              <CalendarSyncIcon />
              Recurring
            </LinkButton>
            <ExternalLinkButton
              to="https://www.getalby.com/topup"
              variant="ghost"
              size="sm"
              className="hidden sm:inline-flex"
            >
              <CreditCardIcon />
              Buy
            </ExternalLinkButton>
            <WalletActionsMenu hasChannelManagement={!!hasChannelManagement} />
          </div>
        }
      />
      <Outlet />
    </>
  );
}
